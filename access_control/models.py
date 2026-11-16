from dataclasses import dataclass, field
from typing import Any, Literal


Decision = Literal['permit', 'deny']


@dataclass
class RuleEvaluation:
    rule_id: str
    result: bool


@dataclass
class AccessDecision:
    decision: Decision

    resource_id: str | None = None
    policy_reference: str | None = None
    policy_version: int | None = None

    evaluations: list[RuleEvaluation] = field(
        default_factory=list
    )

    metadata: dict[str, Any] = field(
        default_factory=dict
    )

    @property
    def granted(self) -> bool:
        return self.decision == 'permit'