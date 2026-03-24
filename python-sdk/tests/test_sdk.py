import json
import os
import subprocess
import sys
import tempfile
import unittest


SDK_PATH = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))


def run_plugin(script_body, stdin_data=None):
    """Write a plugin script to a temp file, run it, and return (stdout, stderr, returncode)."""
    with tempfile.NamedTemporaryFile(mode="w", suffix=".py", delete=False) as f:
        f.write(script_body)
        f.flush()
        try:
            stdin_bytes = None
            if stdin_data is not None:
                stdin_bytes = json.dumps(stdin_data).encode() if isinstance(stdin_data, dict) else stdin_data.encode()
            result = subprocess.run(
                [sys.executable, f.name],
                input=stdin_bytes,
                capture_output=True,
                timeout=10,
            )
            return result.stdout.decode(), result.stderr.decode(), result.returncode
        finally:
            os.unlink(f.name)


def make_plugin(render_body):
    """Build a plugin script string that imports the SDK from the correct path."""
    return f'''
import sys
sys.path.insert(0, "{SDK_PATH}")
from ccstatuswidgets import widget

@widget(name="test")
def render(input_data, config):
{render_body}
'''


class TestWidget(unittest.TestCase):

    def test_widget_produces_correct_output(self):
        plugin = make_plugin('    return {"text": "hello", "color": "green"}')
        stdout, stderr, rc = run_plugin(plugin, stdin_data={})
        self.assertEqual(rc, 0, f"Plugin failed: {stderr}")
        output = json.loads(stdout)
        self.assertEqual(output, {"text": "hello", "color": "green"})

    def test_widget_handles_empty_stdin(self):
        plugin = make_plugin('    return {"text": "empty"}')
        stdout, stderr, rc = run_plugin(plugin, stdin_data="")
        self.assertEqual(rc, 0, f"Plugin failed: {stderr}")
        output = json.loads(stdout)
        self.assertEqual(output, {"text": "empty"})

    def test_widget_passes_input_data(self):
        plugin = make_plugin('    return {"got_key": input_data.get("my_key", "missing")}')
        stdout, stderr, rc = run_plugin(plugin, stdin_data={"my_key": "my_value"})
        self.assertEqual(rc, 0, f"Plugin failed: {stderr}")
        output = json.loads(stdout)
        self.assertEqual(output, {"got_key": "my_value"})

    def test_widget_passes_config(self):
        plugin = make_plugin('    return {"cfg": config.get("setting", "none")}')
        stdin_data = {"_widget_config": {"setting": "enabled"}}
        stdout, stderr, rc = run_plugin(plugin, stdin_data=stdin_data)
        self.assertEqual(rc, 0, f"Plugin failed: {stderr}")
        output = json.loads(stdout)
        self.assertEqual(output, {"cfg": "enabled"})

    def test_widget_config_defaults_to_empty_dict(self):
        plugin = make_plugin('    return {"cfg_type": type(config).__name__, "cfg_len": len(config)}')
        stdout, stderr, rc = run_plugin(plugin, stdin_data={"some": "data"})
        self.assertEqual(rc, 0, f"Plugin failed: {stderr}")
        output = json.loads(stdout)
        self.assertEqual(output, {"cfg_type": "dict", "cfg_len": 0})

    def test_widget_none_return_produces_no_output(self):
        plugin = make_plugin('    return None')
        stdout, stderr, rc = run_plugin(plugin, stdin_data={})
        self.assertEqual(rc, 0, f"Plugin failed: {stderr}")
        self.assertEqual(stdout.strip(), "")

    def test_widget_error_exits_nonzero(self):
        plugin = make_plugin('    raise ValueError("boom")')
        stdout, stderr, rc = run_plugin(plugin, stdin_data={})
        self.assertNotEqual(rc, 0)
        self.assertIn("Plugin error (test)", stderr)
        self.assertIn("boom", stderr)


if __name__ == "__main__":
    unittest.main()
