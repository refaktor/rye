# HN Surf OpenAI - Python Version

Python equivalent of the `hn_surf_openai.rye` script. This script fetches news from Hacker News and Lobsters, then uses OpenAI to summarize the top 3 articles that match your interests.

## Features

- Fetches content from Hacker News and Lobsters
- Uses OpenAI API to analyze articles based on your interests
- Provides streaming output for real-time responses
- Colored terminal output for better readability
- Configurable interests and news sites

## Requirements

- Python 3.7+
- OpenAI API key
- Internet connection

## Installation

1. Install the required dependencies:
```bash
pip install -r requirements.txt
```

2. Create a `.oai-token` file in the same directory with your OpenAI API key:
```bash
echo "your-openai-api-key-here" > .oai-token
```

## Usage

Simply run the script:
```bash
python hn_surf_openai.py
```

Or make it executable and run directly:
```bash
chmod +x hn_surf_openai.py
./hn_surf_openai.py
```

## Configuration

You can customize the script by modifying the configuration variables at the top:

```python
MY_INTERESTS = ["programming languages", "user interfaces"]
MY_SITES = ["https://news.ycombinator.com/", "https://lobste.rs/"]
```

## How it Works

1. **Content Fetching**: The script visits each configured news site and extracts the text content from the page body
2. **AI Analysis**: It sends the content to OpenAI along with your interests
3. **Streaming Output**: OpenAI's response is streamed in real-time to provide immediate feedback
4. **Formatted Display**: Results are displayed with colored headers to distinguish between different sites

## Example Output

```
SITE https://news.ycombinator.com/
Based on your interests in programming languages and user interfaces, here are the top 3 articles that might interest you:

1. **New Rust Framework for UI Development** - A discussion about a new framework...
2. **Python Type Annotations Update** - Latest improvements to Python's type system...
3. **Modern CSS Techniques for Better UX** - Advanced CSS methods for improved user interfaces...

----

SITE https://lobste.rs/
Based on your interests in programming languages and user interfaces, here are the top 3 articles that might interest you:

1. **Functional Programming in Modern JavaScript** - Exploring functional concepts...
2. **Design System Implementation** - Best practices for building design systems...
3. **WebAssembly Performance Comparison** - Benchmarking different language compilations...

----
```

## Dependencies

- `requests`: For HTTP requests to fetch web content
- `beautifulsoup4`: For parsing HTML and extracting text
- `openai`: Official OpenAI Python client
- `colorama`: Cross-platform colored terminal output

## Error Handling

The script includes error handling for:
- Missing or invalid OpenAI API key
- Network connectivity issues
- Invalid responses from news sites
- OpenAI API errors

## Comparison to Original Rye Script

This Python version replicates the functionality of the original `hn_surf_openai.rye` script:

| Feature | Rye Version | Python Version |
|---------|-------------|----------------|
| Content fetching | `surf` module | `requests` + `BeautifulSoup` |
| OpenAI integration | `openai` builtin | `openai` Python client |
| Streaming output | Built-in streaming | OpenAI streaming API |
| Terminal colors | `term/` builtins | `colorama` |
| Token management | `Read %.oai-token` | File reading with error handling |

## License

This script maintains the same functionality as the original Rye version while being implemented in standard Python with popular libraries.
