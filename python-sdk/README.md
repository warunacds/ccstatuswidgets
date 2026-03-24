# ccstatuswidgets Python SDK

Python SDK for building [ccstatuswidgets](https://github.com/warunacds/ccstatuswidgets) plugins.

## Install

```bash
pip install ccstatuswidgets
```

Or install from source:

```bash
cd python-sdk
pip install -e .
```

## Usage

Create a plugin by decorating a render function with `@widget`:

```python
from ccstatuswidgets import widget

@widget(name="my-widget")
def render(input_data, config):
    return {"text": "hello", "color": "green"}
```

The decorator handles the stdin/stdout JSON protocol automatically:

1. Reads JSON from stdin
2. Extracts `_widget_config` from the input (if present) and passes it as `config`
3. Calls your render function with `(input_data, config)`
4. Writes the returned dict as JSON to stdout

### Input

Your render function receives two arguments:

- **`input_data`** -- the full JSON object read from stdin
- **`config`** -- the value of `input_data["_widget_config"]` (or `{}` if absent)

### Output

Return a dict with widget display properties:

```python
{
    "text": "displayed text",
    "color": "green"       # optional
}
```

Return `None` to produce no output.

## Testing locally

```bash
echo '{"_widget_config": {"key": "value"}}' | python3 my_plugin.py
```

## Running the test suite

```bash
cd python-sdk
python3 -m pytest tests/ -v
```
