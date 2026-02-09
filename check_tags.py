#!/usr/bin/env python3
"""Check for unclosed template tags."""

filepath = r'c:\Users\ssebi\Desktop\swadiq-schools\app\templates\teachers\view.html'

with open(filepath, 'r', encoding='utf-8') as f:
    content = f.read()

# Check for unclosed tags {{ without }}
# This is tricky because {{ within string is valid... but rare in Go templates
last_open = 0
while True:
    open_idx = content.find('{{', last_open)
    if open_idx == -1:
        break
    
    close_idx = content.find('}}', open_idx)
    if close_idx == -1:
        print(f"ERROR: Unclosed tag starting at char {open_idx}")
        # Find line number
        line_num = content[:open_idx].count('\n') + 1
        print(f"  Line {line_num}: {content[open_idx:open_idx+50]}...")
        break
    
    last_open = open_idx + 2

print("Finished checking tags.")
