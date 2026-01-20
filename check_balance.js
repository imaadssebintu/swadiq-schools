
const fs = require('fs');
const content = fs.readFileSync('c:/Users/ssebi/Desktop/projects/swadiq-schools/app/templates/events/index.html', 'utf8');

let stack = [];
let i = 0;
while (i < content.length) {
    if (content.substring(i, i + 2) === '{{') {
        stack.push({ type: '{{', pos: i });
        i += 2;
    } else if (content.substring(i, i + 2) === '}}') {
        if (stack.length > 0) {
            stack.pop();
        } else {
            console.log('Extra }} at pos', i);
        }
        i += 2;
    } else {
        i++;
    }
}

if (stack.length > 0) {
    stack.forEach(s => console.log('Unclosed {{ at pos', s.pos));
} else {
    console.log('Balanced!');
}
