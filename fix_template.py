#!/usr/bin/env python3
"""Fix HTML template by removing extra closing divs."""

filepath = r'c:\Users\ssebi\Desktop\swadiq-schools\app\templates\teachers\view.html'

with open(filepath, 'r', encoding='utf-8') as f:
    lines = f.readlines()

# Find lines with extra closing divs
new_lines = []
div_balance = 0
removed_count = 0

for i, line in enumerate(lines, 1):
    opens = line.count('<div')
    closes = line.count('</div>')
    
    # Calculate what balance would be after this line
    new_balance = div_balance + opens - closes
    
    # If balance would go negative, we have extra closing divs
    if new_balance < 0:
        # Remove enough closing divs to keep balance at 0
        closes_to_keep = div_balance + opens
        closes_to_remove = closes - closes_to_keep
        
        if closes_to_remove > 0:
            print(f"Line {i}: Removing {closes_to_remove} extra </div> tags")
            # Remove the extra closing divs
            modified_line = line
            for _ in range(closes_to_remove):
                modified_line = modified_line.replace('</div>', '', 1)
                removed_count += 1
            new_lines.append(modified_line)
            div_balance += opens - closes_to_keep
        else:
            new_lines.append(line)
            div_balance = new_balance
    else:
        new_lines.append(line)
        div_balance = new_balance

print(f"\nTotal extra closing divs removed: {removed_count}")
print(f"Final div balance: {div_balance}")

# Write fixed file
with open(filepath, 'w', encoding='utf-8', newline='') as f:
    f.writelines(new_lines)

print(f"File fixed and saved!")
