#!/usr/bin/env python3
import sys
import json
import os

def main():
    # Log that the script started
    with open("external_handler.log", "a") as f:
        f.write(f"Script started with arguments: {sys.argv}\n")

    # Read stdin
    try:
        input_data = sys.stdin.read()
        with open("external_handler.log", "a") as f:
            f.write(f"Received stdin: {input_data}\n")
        
        if input_data:
            try:
                config = json.loads(input_data)
                with open("external_handler.log", "a") as f:
                    f.write(f"Successfully parsed JSON config. Keys: {list(config.keys())}\n")
            except json.JSONDecodeError as e:
                 with open("external_handler.log", "a") as f:
                    f.write(f"Failed to parse JSON: {e}\n")
    except Exception as e:
        with open("external_handler.log", "a") as f:
            f.write(f"Error reading stdin: {e}\n")

    # Return a dummy category
    print("documents")

if __name__ == "__main__":
    main()
