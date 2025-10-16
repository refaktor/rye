import requests
from bs4 import BeautifulSoup
from openai import OpenAI
from colorama import init, Fore, Style

init()

# Configuration
MY_INTERESTS = ["programming languages", "user interfaces"]
MY_SITES = ["https://news.ycombinator.com/", "https://lobste.rs/"]

def get_content(url):
    headers = {'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'}
    response = requests.get(url, headers=headers, timeout=10)
    soup = BeautifulSoup(response.content, 'html.parser')
    return soup.find('body').get_text(separator=' ', strip=True)

def print_header(site_name):
    print(f"{Fore.CYAN}{Style.BRIGHT}\nSITE {site_name}{Style.RESET_ALL}")

def analyze_with_ai(content, interests):
    with open('.oai-token', 'r') as f:
        api_key = f.read().strip()
    
    client = OpenAI(api_key=api_key)
    
    prompt = f"""My interests are: {', '.join(interests)}

look at the news titles above and summarize top 3 that might interest me

Content:
{content[:8000]}"""
    
    response = client.chat.completions.create(
        messages=[{"role": "user", "content": prompt}],
        stream=True,
    )
    
    for chunk in response:
        if chunk.choices[0].delta.content is not None:
            print(chunk.choices[0].delta.content, end='', flush=True)

def main():
    for site_url in MY_SITES:
        print_header(site_url)
        content = get_content(site_url)
        analyze_with_ai(content, MY_INTERESTS)
        print("\n\n----\n")

if __name__ == "__main__":
    main()
