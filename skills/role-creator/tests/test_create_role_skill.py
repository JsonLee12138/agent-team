import importlib.util
import sys
import tempfile
import unittest
from datetime import datetime
from pathlib import Path


def load_module():
    script_path = (
        Path(__file__).resolve().parents[1] / "scripts" / "create_role_skill.py"
    )
    spec = importlib.util.spec_from_file_location("create_role_skill", script_path)
    module = importlib.util.module_from_spec(spec)
    assert spec.loader is not None
    sys.modules[spec.name] = module
    spec.loader.exec_module(module)
    return module


class RoleCreatorScriptTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.module = load_module()
        cls.templates_dir = Path(__file__).resolve().parents[1] / "assets" / "templates"

    def test_invalid_name_includes_kebab_case_suggestion(self):
        with self.assertRaisesRegex(ValueError, "frontend-dev"):
            self.module.validate_role_name("FrontEnd Dev")

    def test_render_files_is_deterministic(self):
        config = self.module.RoleConfig(
            role_name="frontend-dev",
            description="Frontend role for UI implementation",
            system_goal="Ship accessible and maintainable UI work",
            in_scope=["Implement UI components", "Improve page accessibility"],
            out_of_scope=["Database migrations", "Backend API ownership"],
            skills=["vitest", "ui-ux-pro-max"],
        )
        first = self.module.render_files(config, self.templates_dir)
        second = self.module.render_files(config, self.templates_dir)
        self.assertEqual(first, second)

    def test_overwrite_creates_backup_before_replacing_managed_files(self):
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_root = Path(tmpdir)
            target = repo_root / "skills" / "frontend-dev"
            references = target / "references"
            target.mkdir(parents=True)
            (target / "SKILL.md").write_text("old skill\n", encoding="utf-8")
            (target / "role.yaml").write_text("legacy root yaml\n", encoding="utf-8")
            references.mkdir(parents=True)
            (references / "role.yaml").write_text("old yaml\n", encoding="utf-8")
            (target / "system.md").write_text("old system\n", encoding="utf-8")
            (target / "keep.txt").write_text("keep me\n", encoding="utf-8")

            config = self.module.RoleConfig(
                role_name="frontend-dev",
                description="Frontend role for UI implementation",
                system_goal="Ship accessible and maintainable UI work",
                in_scope=["Implement UI components"],
                out_of_scope=["Own backend services"],
                skills=["vitest"],
            )
            result = self.module.create_or_update_role(
                repo_root=repo_root,
                config=config,
                templates_dir=self.templates_dir,
                overwrite_mode="yes",
                now_fn=lambda: datetime(2026, 2, 25, 12, 0, 0),
            )

            expected_backup = (
                repo_root / "skills" / ".backup" / "frontend-dev-20260225-120000"
            )
            self.assertEqual(result.backup_path, expected_backup)
            self.assertTrue((expected_backup / "SKILL.md").exists())
            self.assertEqual(
                (expected_backup / "references" / "role.yaml").read_text(
                    encoding="utf-8"
                ),
                "old yaml\n",
            )
            self.assertEqual(
                (expected_backup / "role.yaml").read_text(encoding="utf-8"),
                "legacy root yaml\n",
            )
            self.assertEqual(
                (expected_backup / "SKILL.md").read_text(encoding="utf-8"), "old skill\n"
            )
            self.assertEqual(
                (target / "keep.txt").read_text(encoding="utf-8"), "keep me\n"
            )
            self.assertIn(
                "Frontend role for UI implementation",
                (target / "SKILL.md").read_text(encoding="utf-8"),
            )
            self.assertIn(
                "Frontend role for UI implementation",
                (target / "references" / "role.yaml").read_text(encoding="utf-8"),
            )
            self.assertFalse((target / "role.yaml").exists())

    def test_manual_skills_fallback_when_recommendation_empty(self):
        final = self.module.resolve_final_skills(
            selected_skills=[],
            recommended_skills=[],
            added_skills=[],
            removed_skills=[],
            manual_skills=["custom-skill-a", "custom-skill-b"],
        )
        self.assertEqual(final, ["custom-skill-a", "custom-skill-b"])

    def test_empty_skills_allowed(self):
        final = self.module.resolve_final_skills(
            selected_skills=[],
            recommended_skills=[],
            added_skills=[],
            removed_skills=[],
            manual_skills=[],
        )
        self.assertEqual(final, [])

    def test_role_yaml_uses_empty_array_for_empty_skills(self):
        config = self.module.RoleConfig(
            role_name="frontend-dev",
            description="Frontend role",
            system_goal="Ship UI",
            in_scope=["Build components"],
            out_of_scope=["Own backend"],
            skills=[],
        )
        rendered = self.module.render_files(config, self.templates_dir)
        self.assertIn("skills: []", rendered["references/role.yaml"])

    def test_role_yaml_keeps_list_shape_for_non_empty_skills(self):
        config = self.module.RoleConfig(
            role_name="frontend-dev",
            description="Frontend role",
            system_goal="Ship UI",
            in_scope=["Build components"],
            out_of_scope=["Own backend"],
            skills=["vitest", "ui-ux-pro-max"],
        )
        rendered = self.module.render_files(config, self.templates_dir)
        self.assertIn(
            'skills:\n  - "vitest"\n  - "ui-ux-pro-max"',
            rendered["references/role.yaml"],
        )

    def test_parse_selection_checkbox_mode(self):
        selected, mode, checkbox_precedence = self.module.parse_selection_reply(
            "1. [x] ui-ux-pro-max\n2. [ ] vitest\n3. [x] better-icons",
            ["ui-ux-pro-max", "vitest", "better-icons"],
        )
        self.assertEqual(selected, ["ui-ux-pro-max", "better-icons"])
        self.assertEqual(mode, "checkbox")
        self.assertFalse(checkbox_precedence)

    def test_parse_selection_numeric_mode(self):
        selected, mode, checkbox_precedence = self.module.parse_selection_reply(
            "1,3",
            ["ui-ux-pro-max", "vitest", "better-icons"],
        )
        self.assertEqual(selected, ["ui-ux-pro-max", "better-icons"])
        self.assertEqual(mode, "numeric")
        self.assertFalse(checkbox_precedence)

    def test_parse_selection_checkbox_precedence_when_mixed(self):
        selected, mode, checkbox_precedence = self.module.parse_selection_reply(
            "1. [ ] ui-ux-pro-max\n2. [x] vitest\n3. [ ] better-icons\n1,3",
            ["ui-ux-pro-max", "vitest", "better-icons"],
        )
        self.assertEqual(selected, ["vitest"])
        self.assertEqual(mode, "checkbox")
        self.assertTrue(checkbox_precedence)

    def test_create_new_role_happy_path(self):
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_root = Path(tmpdir)
            config = self.module.RoleConfig(
                role_name="data-engineer",
                description="Data pipeline and ETL role",
                system_goal="Build reliable data pipelines",
                in_scope=["ETL jobs", "Data validation"],
                out_of_scope=["Frontend work"],
                skills=["vitest"],
            )
            result = self.module.create_or_update_role(
                repo_root=repo_root,
                config=config,
                templates_dir=self.templates_dir,
                overwrite_mode="ask",
            )
            self.assertIsNone(result.backup_path)
            target = repo_root / "skills" / "data-engineer"
            self.assertTrue((target / "SKILL.md").exists())
            self.assertTrue((target / "references" / "role.yaml").exists())
            self.assertTrue((target / "system.md").exists())
            skill_md = (target / "SKILL.md").read_text(encoding="utf-8")
            self.assertIn("data-engineer", skill_md)
            self.assertIn("Data pipeline and ETL role", skill_md)
            role_yaml = (target / "references" / "role.yaml").read_text(encoding="utf-8")
            self.assertIn("ETL jobs", role_yaml)
            self.assertIn("Frontend work", role_yaml)

    def test_overwrite_mode_no_raises(self):
        with tempfile.TemporaryDirectory() as tmpdir:
            repo_root = Path(tmpdir)
            target = repo_root / "skills" / "existing-role"
            target.mkdir(parents=True)
            (target / "SKILL.md").write_text("old\n", encoding="utf-8")
            config = self.module.RoleConfig(
                role_name="existing-role",
                description="Some role",
                system_goal="Do stuff",
                in_scope=["Task A"],
                out_of_scope=["Task B"],
                skills=[],
            )
            with self.assertRaises(FileExistsError):
                self.module.create_or_update_role(
                    repo_root=repo_root,
                    config=config,
                    templates_dir=self.templates_dir,
                    overwrite_mode="no",
                )

    def test_normalize_role_name_edge_cases(self):
        self.assertEqual(self.module.normalize_role_name("  FrontEnd Dev  "), "frontend-dev")
        self.assertEqual(self.module.normalize_role_name("a---b"), "a-b")
        self.assertEqual(self.module.normalize_role_name("Hello World!@#"), "hello-world")
        self.assertEqual(self.module.normalize_role_name("   "), "")
        self.assertEqual(self.module.normalize_role_name("UPPER_CASE_NAME"), "upper-case-name")
        self.assertEqual(self.module.normalize_role_name("already-kebab"), "already-kebab")

    def test_collect_scope_fallback(self):
        self.assertEqual(
            self.module.collect_scope([], fallback="Default scope"),
            ["Default scope"],
        )
        self.assertEqual(
            self.module.collect_scope(["", "  "], fallback="Default scope"),
            ["Default scope"],
        )
        self.assertEqual(
            self.module.collect_scope(["a,b", "c"], fallback="Default"),
            ["a", "b", "c"],
        )


if __name__ == "__main__":
    unittest.main()
