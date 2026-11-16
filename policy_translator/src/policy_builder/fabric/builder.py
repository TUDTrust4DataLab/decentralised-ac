import json
import re
import time
from pathlib import Path


IDENT = re.compile(r"^[A-Za-z][A-Za-z0-9_-]*$")


class UnsupportedNIR(ValueError):
    pass


def go_name(name: str) -> str:
    if not IDENT.fullmatch(name):
        raise UnsupportedNIR(f"Invalid identifier: {name}")

    return "".join(
        part[:1].upper() + part[1:]
        for part in name.split("-")
    )


def go_variable(name: str) -> str:
    if not IDENT.fullmatch(name):
        raise UnsupportedNIR(f"Invalid variable: {name}")

    return name.replace("-", "_")


def quote(value: str) -> str:
    return json.dumps(value)


def resolve_variable(name: str, scope: dict) -> str:
    if name not in scope:
        raise UnsupportedNIR(
            f"Unbound variable: {name}"
        )

    return scope[name]


def validate_fact_reference(
    node: dict,
    facts: dict,
) -> None:
    name = node["name"]

    if name not in facts:
        raise UnsupportedNIR(
            f"Unknown fact: {name}"
        )

    expected = len(facts[name]["parameters"])
    received = len(node["arguments"])

    if expected != received:
        raise UnsupportedNIR(
            f"Invalid arity for {name}: "
            f"expected {expected}, received {received}"
        )


def expression(
    node: dict,
    scope: dict,
    facts: dict,
) -> str:
    kind = node["type"]

    if kind == "boolean":
        value = str(node["value"]).lower()
        return f"boolResult({value})"

    if kind == "fact_reference":
        validate_fact_reference(node, facts)

        args = ", ".join(
            resolve_variable(arg, scope)
            for arg in node["arguments"]
        )

        return (
            f"evalFact(ctx, {quote(node['name'])}, "
            f"[]string{{{args}}})"
        )

    if kind in ("and", "or"):
        helper = "allOf" if kind == "and" else "anyOf"

        children = []

        for child in node["operands"]:
            generated = expression(
                child, scope, facts
            )

            children.append(
                "func() (bool, error) { "
                f"return {generated} "
                "}"
            )

        return (
            f"{helper}("
            + ", ".join(children)
            + ")"
        )

    if kind == "not":
        inner = expression(
            node["operand"], scope, facts
        )

        return (
            "negate(func() (bool, error) { "
            f"return {inner} "
            "})"
        )

    raise UnsupportedNIR(
        f"Unsupported expression type: {kind}"
    )


def action_arguments(act: dict) -> list[str]:
    names = [
        act["actor"],
        act["recipient"],
        *act["related_to"],
    ]

    if len(names) != len(set(names)):
        raise UnsupportedNIR(
            f"Repeated parameter in action {act['name']}"
        )

    go_names = [go_variable(name) for name in names]

    if len(go_names) != len(set(go_names)):
        raise UnsupportedNIR(
            f"Conflicting Go parameter names "
            f"in action {act['name']}"
        )

    return names


def generate_fact_evaluator(
    fact: dict,
    facts: dict,
) -> str:
    if not fact["derived"]:
        return ""

    name = fact["name"]
    parameters = fact["parameters"]

    scope = {
        parameter: f"args[{index}]"
        for index, parameter in enumerate(parameters)
    }

    condition = expression(
        fact["condition"],
        scope,
        facts,
    )

    return f"""
    case {quote(name)}:
        if len(args) != {len(parameters)} {{
            return false, fmt.Errorf(
                "invalid arguments for {name}",
            )
        }}
        return {condition}
"""



def generate_effect(
    effect: dict,
    scope: dict,
    facts: dict,
) -> str:
    operation = effect["operation"]
    target = effect["target"]

    if target["type"] != "fact_reference":
        raise UnsupportedNIR(
            "Effect target must be a fact reference"
        )

    validate_fact_reference(target, facts)

    name = target["name"]

    if facts[name]["derived"]:
        raise UnsupportedNIR(
            f"Cannot modify derived fact: {name}"
        )

    args = ", ".join(
        resolve_variable(arg, scope)
        for arg in target["arguments"]
    )


    if operation == "create":
        function = (
            "putDuty"
            if facts[name]["kind"] == "duty"
            else "putFact"
        )

    elif operation == "terminate":
        function = (
            "delDuty"
            if facts[name]["kind"] == "duty"
            else "delFact"
        )

    else:
        raise UnsupportedNIR(
            f"Unsupported effect operation: {operation}"
        )

    return (
        f'{function}(ctx, {quote(name)}, '
        f'[]string{{{args}}})'
    )


def generate_act(
    act: dict,
    facts: dict,
) -> str:
    name = go_name(act["name"])
    parameters = action_arguments(act)

    scope = {
        parameter: go_variable(parameter)
        for parameter in parameters
    }

    # Go requires trailing commas in multiline
    # parameter lists.
    signature = "".join(
        f"    {scope[p]} string,\n"
        for p in parameters
    )

    call_args = ", ".join(
        scope[p] for p in parameters
    )

    call_suffix = (
        ", " + call_args if call_args else ""
    )

    condition = (
        expression(
            act["condition"],
            scope,
            facts,
        )
        if act["condition"] is not None
        else "boolResult(true)"
    )

    effect_statements = []

    for effect in act["effects"]:
        generated = generate_effect(
            effect,
            scope,
            facts,
        )

        effect_statements.append(
            f"""
    if err := {generated}; err != nil {{
        return err
    }}
"""
        )

    effects_code = "\n".join(effect_statements)

    return f"""
func (p *PolicyContract) Can{name}(
    ctx contractapi.TransactionContextInterface,
{signature}) (bool, error) {{
    return {condition}
}}

func (p *PolicyContract) {name}(
    ctx contractapi.TransactionContextInterface,
{signature}) error {{
    allowed, err := p.Can{name}(ctx{call_suffix})

    if err != nil {{
        return err
    }}

    if !allowed {{
        return fmt.Errorf(
            "{act['name']}: not authorised",
        )
    }}

{effects_code}

    return nil
}}
"""


GO_RUNTIME = r'''
package main

import (
    "encoding/json"
    "fmt"

    "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type PolicyContract struct {
    contractapi.Contract
}

func factKey(
    name string,
    args []string,
) (string, error) {
    encoded, err := json.Marshal(
        struct {
            Name string   `json:"name"`
            Args []string `json:"args"`
        }{
            Name: name,
            Args: args,
        },
    )

    if err != nil {
        return "", err
    }

    return "normative:" + string(encoded), nil
}

func getStoredFact(
    ctx contractapi.TransactionContextInterface,
    name string,
    args []string,
) (bool, error) {
    key, err := factKey(name, args)

    if err != nil {
        return false, err
    }

    value, err := ctx.GetStub().GetState(key)

    if err != nil {
        return false, err
    }

    return value != nil, nil
}

func putFact(
    ctx contractapi.TransactionContextInterface,
    name string,
    args []string,
) error {
    key, err := factKey(name, args)

    if err != nil {
        return err
    }

    return ctx.GetStub().PutState(
        key,
        []byte("true"),
    )
}

func delFact(
    ctx contractapi.TransactionContextInterface,
    name string,
    args []string,
) error {
    key, err := factKey(name, args)

    if err != nil {
        return err
    }

    return ctx.GetStub().DelState(key)
}


func dutyKey(
    name string,
    args []string,
) (string, error) {
    encoded, err := json.Marshal(
        struct {
            Name string   `json:"name"`
            Args []string `json:"args"`
        }{
            Name: name,
            Args: args,
        },
    )

    if err != nil {
        return "", err
    }

    return "duty:" + string(encoded), nil
}

func getStoredDuty(
    ctx contractapi.TransactionContextInterface,
    name string,
    args []string,
) (bool, error) {
    key, err := dutyKey(name, args)
    if err != nil {
        return false, err
    }

    value, err := ctx.GetStub().GetState(key)
    if err != nil {
        return false, err
    }

    return value != nil, nil
}

func putDuty(
    ctx contractapi.TransactionContextInterface,
    name string,
    args []string,
) error {
    key, err := dutyKey(name, args)
    if err != nil {
        return err
    }

    return ctx.GetStub().PutState(
        key, []byte("true"),
    )
}

func delDuty(
    ctx contractapi.TransactionContextInterface,
    name string,
    args []string,
) error {
    key, err := dutyKey(name, args)
    if err != nil {
        return err
    }

    return ctx.GetStub().DelState(key)
}

func boolResult(
    value bool,
) (bool, error) {
    return value, nil
}

func allOf(
    checks ...func() (bool, error),
) (bool, error) {
    for _, check := range checks {
        value, err := check()

        if err != nil {
            return false, err
        }

        if !value {
            return false, nil
        }
    }

    return true, nil
}

func anyOf(
    checks ...func() (bool, error),
) (bool, error) {
    for _, check := range checks {
        value, err := check()

        if err != nil {
            return false, err
        }

        if value {
            return true, nil
        }
    }

    return false, nil
}

func negate(
    check func() (bool, error),
) (bool, error) {
    value, err := check()

    if err != nil {
        return false, err
    }

    return !value, nil
}

func (p *PolicyContract) HasFact(
    ctx contractapi.TransactionContextInterface,
    name string,
    argsJSON string,
) (bool, error) {
    var args []string

    if err := json.Unmarshal(
        []byte(argsJSON),
        &args,
    ); err != nil {
        return false, err
    }

    return evalFact(ctx, name, args)
}
'''


def generate_bootstrap(facts: dict) -> str:
    allowed_cases = []

    for fact in facts.values():
        if fact["derived"]:
            continue

        name = fact["name"]
        arity = len(fact["parameters"])

        allowed_cases.append(f"""
    case {quote(name)}:
        if len(fact.Args) != {arity} {{
            return fmt.Errorf(
                "invalid arity for {name}",
            )
        }}
""")

    cases = "\n".join(allowed_cases)

    return f"""
type InitialFact struct {{
    Name string   `json:"name"`
    Args []string `json:"args"`
}}

func validateInitialFact(fact InitialFact) error {{
    switch fact.Name {{
{cases}
    default:
        return fmt.Errorf(
            "fact not allowed during bootstrap: %s",
            fact.Name,
        )
    }}

    for _, arg := range fact.Args {{
        if arg == "" {{
            return fmt.Errorf(
                "empty fact argument is not allowed",
            )
        }}
    }}

    return nil
}}

func requireBootstrapAdmin(
    ctx contractapi.TransactionContextInterface,
) error {{
    identity := ctx.GetClientIdentity()

    mspID, err := identity.GetMSPID()
    if err != nil {{
        return err
    }}

    if mspID != "Org1MSP" {{
        return fmt.Errorf("bootstrap requires Org1MSP")
    }}

    cert, err := identity.GetX509Certificate()
    if err != nil {{
        return err
    }}
    if cert == nil {{
        return fmt.Errorf("X.509 identity required")
    }}

    for _, ou := range cert.Subject.OrganizationalUnit {{
        if ou == "admin" {{
            return nil
        }}
    }}

    return fmt.Errorf(
        "bootstrap requires an admin identity",
    )
}}

func (p *PolicyContract) InitFacts(
    ctx contractapi.TransactionContextInterface,
    factsJSON string,
) error {{
    if err := requireBootstrapAdmin(ctx); err != nil {{
        return err
    }}

    const marker = "_normative:bootstrap:v1"

    previous, err := ctx.GetStub().GetState(marker)
    if err != nil {{
        return err
    }}
    if previous != nil {{
        return fmt.Errorf("already initialised")
    }}

    var facts []InitialFact
    if err := json.Unmarshal(
        []byte(factsJSON), &facts,
    ); err != nil {{
        return err
    }}

    if len(facts) == 0 {{
        return fmt.Errorf("empty initial state")
    }}

    // Validate every fact before writing any state.
    for _, fact := range facts {{
        if err := validateInitialFact(fact); err != nil {{
            return err
        }}
    }}

    for _, fact := range facts {{
        if err := putFact(
            ctx, fact.Name, fact.Args,
        ); err != nil {{
            return err
        }}
    }}

    return ctx.GetStub().PutState(
        marker, []byte("true"),
    )
}}
"""


def build_chaincode(nir: dict) -> tuple[str, int]:
    start = time.perf_counter_ns()


    declarations = [
        *nir["facts"],
        *nir.get("duties", []),
    ]

    facts = {
        item["name"]: {
            **item,
            "kind": item["kind"],
        }
        for item in declarations
    }

    if len(facts) != len(declarations):
        raise UnsupportedNIR(
            "Duplicate normative declarations"
        )

    # if len(facts) != len(nir["facts"]):
    #     raise UnsupportedNIR(
    #         "Duplicate fact declarations"
    #     )

    # Validate and generate derived fact evaluators.
    derived_cases = "".join(
        generate_fact_evaluator(fact, facts)
        for fact in nir["facts"]
        if fact["derived"]
    )

    functions = "".join(
        generate_act(act, facts)
        for act in nir["acts"]
    )

    bootstrap = generate_bootstrap({
        name: declaration
        for name, declaration in facts.items()
        if declaration["kind"] == "fact"
    })

    # Persistent facts are explicitly enumerated
    # instead of allowing arbitrary ledger keys.
    persistent_cases ="".join(
        f"""
    case {quote(declaration['name'])}:
        if len(args) != {len(declaration['parameters'])} {{
            return false, fmt.Errorf(
                "invalid arguments for {declaration['name']}",
            )
        }}

        return {
            "getStoredDuty"
            if declaration["kind"] == "duty"
            else "getStoredFact"
        }(ctx, name, args)
"""
        for declaration in declarations
        if not declaration["derived"]
    )

    evaluator = f"""
func evalFact(
    ctx contractapi.TransactionContextInterface,
    name string,
    args []string,
) (bool, error) {{
    switch name {{
{derived_cases}
{persistent_cases}
    default:
        return false, fmt.Errorf(
            "unknown normative fact: %s",
            name,
        )
    }}
}}
"""

    main = """
func main() {
    chaincode, err := contractapi.NewChaincode(
        new(PolicyContract),
    )

    if err != nil {
        panic(err)
    }

    if err := chaincode.Start(); err != nil {
        panic(err)
    }
}
"""

    source = (
        GO_RUNTIME
        + "\n"
        + evaluator
        + "\n"
        + bootstrap
        + "\n"
        + functions
        + "\n"
        + main
    )

    duration_ns = time.perf_counter_ns() - start

    return source, duration_ns


def generate(
    nir_file: Path,
    output_dir: Path,
) -> dict:
    nir = json.loads(
        nir_file.read_text(encoding="utf-8")
    )

    source, duration_ns = build_chaincode(nir)

    output_dir.mkdir(
        parents=True,
        exist_ok=True,
    )

    output = output_dir / "main.go"

    output.write_text(
        source,
        encoding="utf-8",
    )

    return {
        "stage": "chaincode_generation",
        "status": "success",
        "duration_ns": duration_ns,
        "output": str(output),
        "bytes": output.stat().st_size,
        "generated_facts": len(nir["facts"]),
        "generated_actions": len(nir["acts"]),
    }
