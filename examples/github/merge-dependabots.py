import time
import requests
from rich.console import Console

# 1. Configuration
try:
    with open(".apitoken", "r") as f:
        TOK = f.read().strip()
except FileNotFoundError:
    print("Error: .apitoken file not found.")
    exit(1)

BOT = "dependabot[bot]"
REPO = "refaktor/rye"
URL = "https://api.github.com/repos/"
HEADERS = {
    "Authorization": f"Bearer {TOK}",
    "Accept": "application/vnd.github+json"
}

console = Console()

def do_merge(num):
    """Merges a specific PR number using the squash method."""
    merge_url = f"{URL}{REPO}/pulls/{num}/merge"
    data = {"merge_method": "squash"}
    response = requests.put(merge_url, headers=HEADERS, json=data)
    return response.status_code

def main():
    # 2. Fetch Open PRs
    with console.status("[bold blue]Fetching pull requests...") as status:
        response = requests.get(f"{URL}{REPO}/pulls?state=open", headers=HEADERS)
        response.raise_for_status()
        prs = response.json()

    # 3. Filter for the bot
    bot_prs = [pr for pr in prs if pr["user"]["login"] == BOT]
    console.print(f"[green]{len(bot_prs)} pull requests found.")

    # 4. Loop and Merge
    for pr in bot_prs:
        num = pr["number"]
        
        # Merge action with spinner
        with console.status(f"[bold yellow]Merging #{num}...") as status:
            status_code = do_merge(num)
            if status_code == 200:
                console.print(f"Successfully merged #{num} :white_check_mark:")
            else:
                console.print(f"Failed to merge #{num} (Status: {status_code}) :x:")

        # Sleep action with spinner
        with console.status("[bold cyan]Sleeping for rebase...") as status:
            time.sleep(60)

if __name__ == "__main__":
    main()
