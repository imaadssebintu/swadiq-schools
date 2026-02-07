import re
import os

def fix_template():
    file_path = r'app\templates\teachers\view.html'
    
    if not os.path.exists(file_path):
        print(f"Error: File not found at {file_path}")
        return

    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()

    # 1. Fix the malformed range block
    # The pattern looks for the specific malformed line we've been fighting with
    # It handles variations with spaces inside braces
    malformed_pattern = r'\{\{\s*range\s+\$i,\s+\$r\s*:=\s*\.roles\s*\}\}.*?\{\{\s*end\s*\}\}'
    
    # The correct replacement string
    correct_replacement = "{{ range $i, $r := .roles }}{{ if $i }},{{ end }}'{{$r.Name}}'{{ end }}"
    
    # We'll use a more specific search to target the array initialization
    # Looking for: roles: [ ... ]
    roles_section_pattern = r'(roles:\s*\[\s*)(.*?)(\s*\])'
    
    def replace_roles(match):
        prefix = match.group(1)
        suffix = match.group(3)
        return prefix + correct_replacement + suffix

    new_content = re.sub(roles_section_pattern, replace_roles, content, flags=re.DOTALL)
    
    if new_content != content:
        print("Fixed malformed 'roles' array initialization.")
        content = new_content
    else:
        print(" 'roles' array initialization appeared correct or not found.")

    # 2. Check for missing {{ end }} at the end of the file
    # Count open and close tags to determine if we need to append one
    # Note: simple counting might get confused by comments, but for this specific file it should work
    
    # Count opening blocks
    open_blocks = len(re.findall(r'\{\{\s*(?:define|if|range|with|block)\s', content))
    # Count closing blocks
    close_blocks = len(re.findall(r'\{\{\s*end\s*\}\}', content))
    
    print(f"Tag Analysis - Open: {open_blocks}, Close: {close_blocks}")
    
    if open_blocks > close_blocks:
        missing_count = open_blocks - close_blocks
        print(f"Detected {missing_count} missing '{{{{ end }}}}' tag(s). Appending...")
        
        # Append the missing end tags
        content = content.rstrip() + ('\n{{ end }}' * missing_count)
    elif close_blocks > open_blocks:
        print(f"Warning: Found more closing tags ({close_blocks}) than opening tags ({open_blocks}).")
        # Optional: Remove extra closing tags if they are at the end
        diff = close_blocks - open_blocks
        if content.strip().endswith('{{ end }}' * diff):
             print(f"Removing {diff} extra '{{{{ end }}}}' tag(s) from end of file.")
             content = content.rstrip()
             for _ in range(diff):
                 content = content[:content.rfind('{{ end }}')].rstrip()
    else:
        print("Template tags appear balanced.")

    # Write the fixed content back
    with open(file_path, 'w', encoding='utf-8') as f:
        f.write(content)
    
    print("File updated successfully.")

if __name__ == "__main__":
    fix_template()
