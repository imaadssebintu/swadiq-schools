#!/usr/bin/env python3
"""Fix malformed template tags on line 982."""

filepath = r'c:\Users\ssebi\Desktop\swadiq-schools\app\templates\teachers\view.html'

with open(filepath, 'r', encoding='utf-8') as f:
    lines = f.readlines()

# Line 982 (index 981)
target_idx = 981
expected_content = "        roles: [{{ range $i, $r:= .roles }}{{ if $i }}, { { end } } '{{$r.Name}}'{ { end } }],"

current_content = lines[target_idx].strip()
print(f"Current content at line 982: {current_content}")

if "roles: [" in current_content and "{ { end } }" in current_content:
    print("Found malformed line. Fixing...")
    lines[target_idx] = "        roles: [{{range $i, $r := .roles}}{{if $i}}, {{end}}'{{$r.Name}}'{{end}}],\n"
    
    with open(filepath, 'w', encoding='utf-8', newline='') as f:
        f.writelines(lines)
    print("File updated successfully.")
else:
    print("Target line not found or already fixed.")
    # Search for it
    for i, line in enumerate(lines):
        if "roles: [" in line and "{ { end } }" in line:
            print(f"Found it at line {i+1}. Fixing...")
            lines[i] = "        roles: [{{range $i, $r := .roles}}{{if $i}}, {{end}}'{{$r.Name}}'{{end}}],\n"
            with open(filepath, 'w', encoding='utf-8', newline='') as f:
                f.writelines(lines)
            print("File updated successfully.")
            break
