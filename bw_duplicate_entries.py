#!/usr/bin/env python3
"""Load and analyze Bitwarden export for duplicate entries."""

import argparse
import json
import os
import pathlib
from pprint import pprint
from dotenv import load_dotenv

# TODO: Incorporate listing of N oldest items.
#
# $ bw login
# ? Email address: dstautberg@gmail.com
# ? Master password: [hidden]
# You are logged in!

# To unlock your vault, set your session key to the `BW_SESSION` environment variable. ex:
# $ export BW_SESSION="nYaQXQsWYOJs8Dpcwy6GSnEsVytJnUb+7W+Z25SqlN1DTxVoYppz0gsjYXZK27by816/Yl8nNnXjSLd9/wpC0Q=="
# > $env:BW_SESSION="nYaQXQsWYOJs8Dpcwy6GSnEsVytJnUb+7W+Z25SqlN1DTxVoYppz0gsjYXZK27by816/Yl8nNnXjSLd9/wpC0Q=="

# You can also pass the session key to any command with the `--session` option. ex:
# $ bw list items --session nYaQXQsWYOJs8Dpcwy6GSnEsVytJnUb+7W+Z25SqlN1DTxVoYppz0gsjYXZK27by816/Yl8nNnXjSLd9/wpC0Q==

# $ bw sync
# $ bw list items | jq 'sort_by(.revisionDate) | .[0:5]'
 
def load_bitwarden_export(file_path: str = "bitwarden_export.json") -> dict:
    """Load the Bitwarden export JSON file and return as Python objects.
    
    Args:
        file_path: Path to the Bitwarden export JSON file
        
    Returns:
        Parsed JSON data as a dictionary
    """
    path = pathlib.Path(file_path)
    
    if not path.exists():
        raise FileNotFoundError(f"File not found: {path}")
    
    with open(path, "r", encoding="utf-8") as f:
        data = json.load(f)
    
    return data


def main():
    """Load and print the Bitwarden export."""
    # Parse command line arguments
    parser = argparse.ArgumentParser(description="Analyze Bitwarden export for duplicate entries")
    parser.add_argument("export_file", nargs="?", default="bitwarden_export.json",
                        help="Path to Bitwarden export JSON file")
    args = parser.parse_args()
    
    # Load environment variables from .env file
    load_dotenv()
    
    # Load search strings from .env file
    search_strings_env = os.getenv('SEARCH_STRINGS', 'password,1234,secret')
    search_strings = [s.strip() for s in search_strings_env.split(',')]
    red_color = '\033[91m'
    reset_color = '\033[0m'
    
    try:
        data = load_bitwarden_export(args.export_file)
        print('-'*50)
        print(f"Bitwarden export loaded successfully: {args.export_file}")
        print(f"Data structure: {type(data)}")
        print(f"Top-level keys: {list(data.keys()) if isinstance(data, dict) else 'N/A'}")
        
        items = data.get("items", [])
        print(f"Number of total items: {len(items)}")
        
        # Count name occurrences
        name_counts = {}
        for item in items:
            name = item.get('name')
            if name:
                name_counts[name] = name_counts.get(name, 0) + 1
        
        # Filter to only items with duplicate names
        duplicate_items = [item for item in items if name_counts.get(item.get('name'), 0) > 1]
        print(f"Number of items with duplicate names: {len(duplicate_items)}")
        print('-'*50)
        
        for item in duplicate_items:
            item_id = item.get('id')
            name = item.get('name')
            username = item.get('login', {}).get('username')
            password = item.get('login', {}).get('password') or ''
            
            # Check if password contains any search strings
            contains_search = any(search_str.lower() in password.lower() for search_str in search_strings)
            
            output = f"Item ID: {item_id}, Name: {name}, username: {username}, password: {password}"
            if contains_search:
                output = f"{red_color}{output}{reset_color}"
            
            print(output)
        # print("\nFull data:")
        # pprint(data)
    except Exception as e:
        print(f"Error: {e}")


if __name__ == "__main__":
    main()
