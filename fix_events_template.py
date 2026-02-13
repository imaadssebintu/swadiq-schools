import re

file_path = r'c:\Users\ssebi\Desktop\swadiq-schools\app\templates\events\index.html'

def fix_template():
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        # Patterns to fix: { { range.Events } } -> {{range .Events}}
        #                  { { end } } -> {{end}}
        # General clean up of double braces with spaces inside
        
        # Fix specific range tag
        new_content = re.sub(r'\{\s*\{\s*range\.Events\s*\}\s*\}', '{{range .Events}}', content)
        
        # Fix specific end tag
        new_content = re.sub(r'\{\s*\{\s*end\s*\}\s*\}', '{{end}}', new_content)
        
        # Generalized fix for other potential issues (conservative)
        # new_content = re.sub(r'\{\s+\{', '{{', new_content)
        # new_content = re.sub(r'\}\s+\}', '}}', new_content)

        if content != new_content:
            with open(file_path, 'w', encoding='utf-8') as f:
                f.write(new_content)
            print("Successfully fixed template tags.")
        else:
            print("No changes needed. Template appears correct.")
            
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    fix_template()
