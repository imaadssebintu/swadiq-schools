#!/usr/bin/env python3
"""Validate Go template structure."""

import re

filepath = r'c:\Users\ssebi\Desktop\swadiq-schools\app\templates\teachers\view.html'

with open(filepath, 'r', encoding='utf-8') as f:
    content = f.read()

# Check for unclosed comments
open_comments = content.count('<!--')
close_comments = content.count('-->')
print(f"Comments: {open_comments} open, {close_comments} close")

if open_comments > close_comments:
    print("ERROR: Unclosed HTML comment found!")
    # Find the last unclosed comment
    last_open = content.rfind('<!--')
    if content.find('-->', last_open) == -1:
        line_num = content[:last_open].count('\n') + 1
        print(f"  Starts at line {line_num}")

# Check for unclosed blocks
# Primitive check for now
block_stack = []
lines = content.split('\n')

for i, line in enumerate(lines, 1):
    # Find all tags
    tags = re.findall(r'\{\{(if|range|with|block|define|end)', line)
    for tag in tags:
        if tag == 'end':
            if not block_stack:
                print(f"Line {i}: Create {{end}} with no opening block")
            else:
                block_stack.pop()
        else:
            block_stack.append((tag, i))

if block_stack:
    print(f"ERROR: Unclosed blocks found: {len(block_stack)}")
    for tag, line in block_stack:
        print(f"  Unclosed {tag} from line {line}")
else:
    print("Blocks are balanced.")

# Check for backticks (Go uses them for raw strings sometimes)
backticks = content.count('`')
if backticks % 2 != 0:
    print(f"ERROR: Odd number of backticks found: {backticks}")

# Check for unclosed JS strings/template literals in script block
script_start = content.find('<script>')
script_end = content.find('</script>')
if script_start != -1 and script_end != -1:
    script_content = content[script_start:script_end]
    # Simple check for quote balance
    single_quotes = script_content.count("'") - script_content.count("\\'")
    double_quotes = script_content.count('"') - script_content.count('\\"')
    back_quotes = script_content.count('`') - script_content.count('\\`')
    
    print(f"Script quotes: '={single_quotes}, \"={double_quotes}, `={back_quotes}")

