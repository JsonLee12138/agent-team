#!/usr/bin/env python3
"""Generate role skill packages under skills/<role-name>/."""

from __future__ import annotations

import argparse
import json
import re
import shutil
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path
from string import Template
from typing import Callable

MANAGED_FILE_TEMPLATES = {
    "SKILL.md": "SKILL.md.tmpl",
    "references/role.yaml": "role.yaml.tmpl",
    "system.md": "system.md.tmpl",
}
KEBAB_CASE_PATTERN = re.compile(r"^[a-z0-9]+(?:-[a-z0-9]+)*$")
CHECKBOX_LINE_PATTERN = re.compile(r"^\s*(\d+)\.\s*\[([xX ])\]")
NUMERIC_SELECTION_PATTERN = re.compile(r"^\d+(?:\s*,\s*\d+)*$")
SCRIPT_DIR = Path(__file__).resolve().parent
DEFAULT_TEMPLATE_DIR = SCRIPT_DIR.parent / "assets" / "templates"


@dataclass(frozen=True)
class RoleConfig:
    role_name: str
    description: str
    system_goal: str
    in_scope: list[str]
    out_of_scope: list[str]
    skills: list[str]


@dataclass(frozen=True)
class GenerationResult:
    target_dir: Path
    backup_path: Path | None


def parse_csv_list(value: str) -> list[str]:
    if not value:
        return []
    return [item.strip() for item in value.split(",") if item.strip()]


def dedupe_keep_order(items: list[str]) -> list[str]:
    seen: set[str] = set()
    ordered: list[str] = []
    for item in items:
        if item not in seen:
            seen.add(item)
            ordered.append(item)
    return ordered


def parse_checkbox_indices(reply: str, max_index: int) -> list[int] | None:
    has_checkbox_lines = False
    checked_indices: list[int] = []
    for raw_line in reply.splitlines():
        match = CHECKBOX_LINE_PATTERN.match(raw_line.strip())
        if match is None:
            continue
        has_checkbox_lines = True
        index = int(match.group(1))
        mark = match.group(2).lower()
        if mark != "x":
            continue
        if 1 <= index <= max_index and index not in checked_indices:
            checked_indices.append(index)
    return checked_indices if has_checkbox_lines else None


def parse_numeric_indices(reply: str, max_index: int) -> list[int]:
    indices: list[int] = []
    for raw_line in reply.splitlines():
        stripped = raw_line.strip()
        if not stripped:
            continue
        if CHECKBOX_LINE_PATTERN.match(stripped):
            continue
        if not NUMERIC_SELECTION_PATTERN.fullmatch(stripped):
            continue
        for part in stripped.split(","):
            index = int(part.strip())
            if 1 <= index <= max_index and index not in indices:
                indices.append(index)
    return indices


def parse_selection_reply(
    reply: str, recommended_skills: list[str]
) -> tuple[list[str] | None, str | None, bool]:
    checkbox_indices = parse_checkbox_indices(reply, max_index=len(recommended_skills))
    numeric_indices = parse_numeric_indices(reply, max_index=len(recommended_skills))

    if checkbox_indices is not None:
        selected = [recommended_skills[index - 1] for index in checkbox_indices]
        return selected, "checkbox", bool(numeric_indices)

    if numeric_indices:
        selected = [recommended_skills[index - 1] for index in numeric_indices]
        return selected, "numeric", False

    return None, None, False


def is_kebab_case(value: str) -> bool:
    return bool(KEBAB_CASE_PATTERN.fullmatch(value))


def normalize_role_name(value: str) -> str:
    lowered = value.strip().lower()
    normalized = re.sub(r"[^a-z0-9]+", "-", lowered)
    normalized = re.sub(r"-{2,}", "-", normalized)
    return normalized.strip("-")


def validate_role_name(role_name: str) -> str:
    if is_kebab_case(role_name):
        return role_name
    suggestion = normalize_role_name(role_name)
    if not suggestion:
        suggestion = "role-name"
    raise ValueError(
        f"role name '{role_name}' must be kebab-case (example: 'frontend-dev'). "
        f"Try '{suggestion}'."
    )


def resolve_final_skills(
    selected_skills: list[str],
    recommended_skills: list[str],
    added_skills: list[str],
    removed_skills: list[str],
    manual_skills: list[str],
) -> list[str]:
    selected = dedupe_keep_order(selected_skills)
    recommended = dedupe_keep_order(recommended_skills)
    manual = dedupe_keep_order(manual_skills)
    added = dedupe_keep_order(added_skills)
    removed = set(dedupe_keep_order(removed_skills))

    base = selected or recommended or manual

    final: list[str] = [skill for skill in base if skill not in removed]
    for skill in added:
        if skill not in removed and skill not in final:
            final.append(skill)
    return final


def md_bullets(items: list[str]) -> str:
    return "\n".join(f"- {item}" for item in items)


def yaml_quote(value: str) -> str:
    return json.dumps(value, ensure_ascii=False)


def yaml_list(items: list[str], indent: int = 0) -> str:
    prefix = " " * indent
    return "\n".join(f"{prefix}- {yaml_quote(item)}" for item in items)


def render_skills_field(skills: list[str]) -> str:
    if not skills:
        return "skills: []"
    return "skills:\n" + yaml_list(skills, indent=2)


def load_template(templates_dir: Path, name: str) -> Template:
    path = templates_dir / name
    if not path.exists():
        raise FileNotFoundError(f"template not found: {path}")
    return Template(path.read_text(encoding="utf-8"))


def render_files(config: RoleConfig, templates_dir: Path) -> dict[str, str]:
    substitutions = {
        "role_name": config.role_name,
        "description": config.description,
        "description_yaml": yaml_quote(config.description),
        "system_goal": config.system_goal,
        "in_scope_md": md_bullets(config.in_scope),
        "out_of_scope_md": md_bullets(config.out_of_scope),
        "in_scope_summary": ", ".join(config.in_scope).lower(),
        "skills_md": md_bullets(config.skills),
        "in_scope_yaml": yaml_list(config.in_scope, indent=4),
        "out_of_scope_yaml": yaml_list(config.out_of_scope, indent=4),
        "skills_field": render_skills_field(config.skills),
    }

    rendered: dict[str, str] = {}
    for output_path, template_name in MANAGED_FILE_TEMPLATES.items():
        template = load_template(templates_dir, template_name)
        content = template.substitute(substitutions).rstrip() + "\n"
        rendered[output_path] = content
    return rendered


def confirm_overwrite(target_dir: Path, input_fn: Callable[[str], str]) -> bool:
    answer = input_fn(
        f"Role directory '{target_dir}' already exists. Overwrite managed files? [y/N]: "
    )
    return answer.strip().lower() in {"y", "yes"}


def backup_existing_role(
    target_dir: Path, role_name: str, now_fn: Callable[[], datetime]
) -> Path:
    backup_root = target_dir.parent / ".backup"
    backup_root.mkdir(parents=True, exist_ok=True)
    timestamp = now_fn().strftime("%Y%m%d-%H%M%S")
    backup_path = backup_root / f"{role_name}-{timestamp}"
    shutil.copytree(target_dir, backup_path)
    return backup_path


def create_or_update_role(
    repo_root: Path,
    config: RoleConfig,
    templates_dir: Path = DEFAULT_TEMPLATE_DIR,
    overwrite_mode: str = "ask",
    input_fn: Callable[[str], str] = input,
    now_fn: Callable[[], datetime] = datetime.now,
) -> GenerationResult:
    validate_role_name(config.role_name)

    skills_root = repo_root / "skills"
    target_dir = skills_root / config.role_name
    backup_path: Path | None = None

    skills_root.mkdir(parents=True, exist_ok=True)
    if target_dir.exists():
        if not target_dir.is_dir():
            raise FileExistsError(f"target path exists and is not a directory: {target_dir}")

        should_overwrite = overwrite_mode == "yes"
        if overwrite_mode == "ask":
            should_overwrite = confirm_overwrite(target_dir, input_fn)
        if overwrite_mode == "no" or not should_overwrite:
            raise FileExistsError(
                f"role directory already exists and overwrite not confirmed: {target_dir}"
            )
        backup_path = backup_existing_role(target_dir, config.role_name, now_fn)
    else:
        target_dir.mkdir(parents=True, exist_ok=True)

    rendered = render_files(config, templates_dir)
    for output_path, content in rendered.items():
        file_path = target_dir / output_path
        file_path.parent.mkdir(parents=True, exist_ok=True)
        file_path.write_text(content, encoding="utf-8")

    legacy_role_yaml = target_dir / "role.yaml"
    if legacy_role_yaml.exists() and legacy_role_yaml.is_file():
        legacy_role_yaml.unlink()

    return GenerationResult(target_dir=target_dir, backup_path=backup_path)


def collect_scope(values: list[str], fallback: str) -> list[str]:
    merged: list[str] = []
    for value in values:
        merged.extend(parse_csv_list(value))
    return merged or [fallback]


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description="Create or update a role skill package under skills/<role-name>/."
    )
    parser.add_argument("--repo-root", default=".", help="Repository root path")
    parser.add_argument("--role-name", required=True, help="Role directory name in kebab-case")
    parser.add_argument("--description", required=True, help="Role description")
    parser.add_argument(
        "--system-goal", required=True, help="Primary objective for the role system prompt"
    )
    parser.add_argument(
        "--in-scope",
        action="append",
        default=[],
        help="In-scope item (repeatable, supports comma-separated values)",
    )
    parser.add_argument(
        "--out-of-scope",
        action="append",
        default=[],
        help="Out-of-scope item (repeatable, supports comma-separated values)",
    )
    parser.add_argument(
        "--skills",
        default="",
        help="Final selected skills (comma-separated). Overrides recommended/manual base.",
    )
    parser.add_argument(
        "--recommended-skills",
        default="",
        help="Recommended skills from find-skills (comma-separated).",
    )
    parser.add_argument("--add-skills", default="", help="Skills to add (comma-separated).")
    parser.add_argument(
        "--remove-skills", default="", help="Skills to remove from the candidate list."
    )
    parser.add_argument(
        "--manual-skills",
        default="",
        help="Manual fallback when recommendation is unavailable or empty.",
    )
    parser.add_argument(
        "--overwrite",
        choices=("ask", "yes", "no"),
        default="ask",
        help="How to handle existing role directory",
    )
    return parser


def main() -> int:
    parser = build_parser()
    args = parser.parse_args()

    try:
        role_name = validate_role_name(args.role_name)
        final_skills = resolve_final_skills(
            selected_skills=parse_csv_list(args.skills),
            recommended_skills=parse_csv_list(args.recommended_skills),
            added_skills=parse_csv_list(args.add_skills),
            removed_skills=parse_csv_list(args.remove_skills),
            manual_skills=parse_csv_list(args.manual_skills),
        )
    except ValueError as exc:
        parser.error(str(exc))

    config = RoleConfig(
        role_name=role_name,
        description=args.description.strip(),
        system_goal=args.system_goal.strip(),
        in_scope=collect_scope(args.in_scope, fallback=args.description.strip()),
        out_of_scope=collect_scope(
            args.out_of_scope, fallback="Tasks outside this role responsibilities"
        ),
        skills=final_skills,
    )

    result = create_or_update_role(
        repo_root=Path(args.repo_root).resolve(),
        config=config,
        templates_dir=DEFAULT_TEMPLATE_DIR,
        overwrite_mode=args.overwrite,
    )

    print(f"Generated role skill at {result.target_dir}")
    if result.backup_path is not None:
        print(f"Backup created at {result.backup_path}")
    print("Managed files: SKILL.md, references/role.yaml, system.md")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
