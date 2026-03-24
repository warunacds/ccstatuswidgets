import json
import sys


def widget(name):
    """Decorator that wraps a render function into a ccstatuswidgets plugin.

    Reads JSON from stdin, passes it to the decorated function along with
    any widget config, and writes the result as JSON to stdout.

    Usage:
        @widget(name="my-widget")
        def render(input_data, config):
            return {"text": "hello", "color": "green"}
    """
    def decorator(func):
        try:
            raw = sys.stdin.read()
            input_data = json.loads(raw) if raw.strip() else {}
            config = input_data.get("_widget_config", {})
            result = func(input_data, config)
            if result is not None:
                json.dump(result, sys.stdout)
                sys.stdout.flush()
        except Exception as e:
            print(f"Plugin error ({name}): {e}", file=sys.stderr)
            sys.exit(1)
        return func
    return decorator
