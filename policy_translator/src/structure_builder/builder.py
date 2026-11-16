import re
import time
from pathlib import Path


class UnsupportedEFlintConstruct(ValueError):
    pass


IDENT = r"[A-Za-z][A-Za-z0-9_-]*"

TOKEN_RE = re.compile(
    r"\s*(&&|\|\||[(),!]|\bTrue\b|\bFalse\b|"
    r"[A-Za-z][A-Za-z0-9_-]*)"
)

DECL_RE = re.compile(
    r"^(Fact|Placeholder|Act|Duty|Event)\s+"
    r"(" + IDENT + r")\b"
)

CLAUSES = (
    "Identified by",
    "Related to",
    "Holds when",
    "Actor",
    "Recipient",
    "Creates",
    "Terminates",
    "Holder",
    "Claimant",
)


def tokenize(expression):
    tokens = []
    pos = 0

    while pos < len(expression):
        if expression[pos:].strip() == "":
            break

        match = TOKEN_RE.match(expression, pos)

        if not match:
            raise UnsupportedEFlintConstruct(
                f"Unsupported expression at: "
                f"{expression[pos:]}"
            )

        tokens.append(match.group(1))
        pos = match.end()

    return tokens


class ExpressionParser:
    def __init__(self, source):
        self.tokens = tokenize(source)
        self.pos = 0

    def peek(self):
        if self.pos >= len(self.tokens):
            return None
        return self.tokens[self.pos]

    def take(self, expected=None):
        token = self.peek()

        if token is None:
            raise ValueError("Unexpected end of expression")

        if expected is not None and token != expected:
            raise ValueError(
                f"Expected {expected!r}, found {token!r}"
            )

        self.pos += 1
        return token

    def parse(self):
        node = self.parse_or()

        if self.peek() is not None:
            raise UnsupportedEFlintConstruct(
                f"Unexpected token: {self.peek()}"
            )

        return node

    def parse_or(self):
        nodes = [self.parse_and()]

        while self.peek() == "||":
            self.take("||")
            nodes.append(self.parse_and())

        return (
            nodes[0] if len(nodes) == 1
            else {"type": "or", "operands": nodes}
        )

    def parse_and(self):
        nodes = [self.parse_unary()]

        while self.peek() == "&&":
            self.take("&&")
            nodes.append(self.parse_unary())

        return (
            nodes[0] if len(nodes) == 1
            else {"type": "and", "operands": nodes}
        )

    def parse_unary(self):
        if self.peek() == "!":
            self.take("!")
            return {
                "type": "not",
                "operand": self.parse_unary(),
            }

        if self.peek() == "(":
            self.take("(")
            node = self.parse_or()
            self.take(")")
            return node

        name = self.take()

        if name in ("True", "False"):
            return {
                "type": "boolean",
                "value": name == "True",
            }

        if not re.fullmatch(IDENT, name):
            raise UnsupportedEFlintConstruct(name)

        if self.peek() == "(":
            self.take("(")
            args = []

            while self.peek() != ")":
                arg = self.take()

                if not re.fullmatch(IDENT, arg):
                    raise UnsupportedEFlintConstruct(arg)

                args.append(arg)

                if self.peek() == ",":
                    self.take(",")
                elif self.peek() != ")":
                    raise ValueError("Expected ',' or ')'")

            self.take(")")

            return {
                "type": "fact_reference",
                "name": name,
                "arguments": args,
            }

        return {
            "type": "variable",
            "name": name,
        }


def parse_expression(value):
    return ExpressionParser(value).parse()


def split_declarations(source):
    declarations = []
    current = None

    for number, line in enumerate(
        source.splitlines(), start=1
    ):
        stripped = line.strip()

        if not stripped:
            continue

        match = DECL_RE.match(stripped)

        if match:
            if current:
                declarations.append(current)

            if match.group(1) == "Event":
                raise UnsupportedEFlintConstruct(
                    f"Unsupported declaration at "
                    f"line {number}: {stripped}"
                )

            current = {
                "kind": match.group(1),
                "name": match.group(2),
                "line": number,
                "body": [],
            }

            remainder = stripped[match.end():].strip()
            if remainder:
                current["body"].append(remainder)

        elif current:
            current["body"].append(stripped)
        else:
            raise UnsupportedEFlintConstruct(
                f"Unexpected content at line {number}"
            )

    if current:
        declarations.append(current)

    return declarations


def parse_clauses(lines):
    result = {}
    key = None

    for line in lines:
        matched = False

        for clause in CLAUSES:
            if (
                line == clause
                or line.startswith(clause + " ")
            ):
                key = clause
                result.setdefault(key, [])
                result[key].append(
                    line[len(clause):].strip()
                )
                matched = True
                break

        if not matched:
            if key is None:
                raise UnsupportedEFlintConstruct(
                    f"Unrecognised clause: {line}"
                )
            result[key].append(line)

    return {
        name: " ".join(parts).strip()
        for name, parts in result.items()
    }


def parse_fact(declaration):
    clauses = parse_clauses(declaration["body"])

    allowed = {"Identified by", "Holds when"}
    if set(clauses) - allowed:
        raise UnsupportedEFlintConstruct(
            f"Unsupported fact clauses: {set(clauses)-allowed}"
        )

    parameters = [
        item.strip()
        for item in clauses.get(
            "Identified by", ""
        ).split("*")
        if item.strip()
    ]

    return {
        "kind": "fact",
        "name": declaration["name"],
        "parameters": parameters,
        "derived": "Holds when" in clauses,
        "condition": (
            parse_expression(clauses["Holds when"])
            if "Holds when" in clauses else None
        ),
        "source_line": declaration["line"],
    }


def parse_duty(declaration):
    clauses = parse_clauses(declaration["body"])

    allowed = {
        "Holder",
        "Claimant",
        "Related to",
    }

    unsupported = set(clauses) - allowed

    if unsupported:
        raise UnsupportedEFlintConstruct(
            f"Unsupported duty clauses: {unsupported}"
        )

    for required in ("Holder", "Claimant"):
        if required not in clauses:
            raise UnsupportedEFlintConstruct(
                f"Missing {required} in "
                f"{declaration['name']}"
            )

    holder = clauses["Holder"]
    claimant = clauses["Claimant"]

    related_to = [
        value.strip()
        for value in clauses.get(
            "Related to", ""
        ).split(",")
        if value.strip()
    ]

    parameters = [
        holder,
        claimant,
        *related_to,
    ]

    if len(parameters) != len(set(parameters)):
        raise UnsupportedEFlintConstruct(
            f"Repeated duty parameter: "
            f"{declaration['name']}"
        )

    return {
        "kind": "duty",
        "name": declaration["name"],
        "holder": holder,
        "claimant": claimant,
        "related_to": related_to,
        "parameters": parameters,
        "derived": False,
        "condition": None,
        "source_line": declaration["line"],
    }


def parse_act(declaration):
    clauses = parse_clauses(declaration["body"])

    allowed = {
        "Actor",
        "Recipient",
        "Related to",
        "Creates",
        "Terminates",
        "Holds when",
    }

    unsupported = set(clauses) - allowed

    if unsupported:
        raise UnsupportedEFlintConstruct(
            f"Unsupported action clauses: {unsupported}"
        )

    for required in ("Actor", "Recipient"):
        if required not in clauses:
            raise UnsupportedEFlintConstruct(
                f"Missing {required} in "
                f"{declaration['name']}"
            )

    effects = []

    for clause, operation in (
        ("Creates", "create"),
        ("Terminates", "terminate"),
    ):
        if clause in clauses:
            target = parse_expression(
                clauses[clause]
            )

            if target["type"] != "fact_reference":
                raise UnsupportedEFlintConstruct(
                    f"{clause} must reference a fact"
                )

            effects.append({
                "operation": operation,
                "target": target,
            })

    return {
        "kind": "act",
        "name": declaration["name"],
        "actor": clauses["Actor"],
        "recipient": clauses["Recipient"],
        "related_to": [
            value.strip()
            for value in clauses.get(
                "Related to", ""
            ).split(",")
            if value.strip()
        ],
        "condition": (
            parse_expression(
                clauses["Holds when"]
            )
            if "Holds when" in clauses
            else None
        ),
        "effects": effects,
        "source_line": declaration["line"],
    }


def build_nir(source):
    start = time.perf_counter_ns()

    declarations = split_declarations(source)

    nir = {
        "version": "0.2",
        "facts": [],
        "duties": [],
        "acts": [],
        "placeholders": [],
    }

    for declaration in declarations:
        kind = declaration["kind"]

        if kind == "Fact":
            nir["facts"].append(
                parse_fact(declaration)
            )

        elif kind == "Duty":
            nir["duties"].append(
                parse_duty(declaration)
            )

        elif kind == "Act":
            nir["acts"].append(
                parse_act(declaration)
            )

        elif kind == "Placeholder":
            body = " ".join(declaration["body"])

            match = re.fullmatch(
                r"For\s+(" + IDENT + r")",
                body,
            )

            if not match:
                raise UnsupportedEFlintConstruct(
                    f"Invalid placeholder: "
                    f"{declaration['name']}"
                )

            nir["placeholders"].append({
                "name": declaration["name"],
                "type": match.group(1),
                "source_line": declaration["line"],
            })

    duration_ns = time.perf_counter_ns() - start

    return nir, duration_ns
