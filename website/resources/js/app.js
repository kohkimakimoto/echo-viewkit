import Alpine from 'alpinejs';
import hljs from 'highlight.js/lib/core';
import django from 'highlight.js/lib/languages/django';
import xml from 'highlight.js/lib/languages/xml';
import go from 'highlight.js/lib/languages/go';
import shell from 'highlight.js/lib/languages/shell';
import javascript from 'highlight.js/lib/languages/javascript';
import 'highlight.js/styles/atom-one-dark.css';

// init alpinejs
window.Alpine = Alpine;
Alpine.start();

// init highlight.js
hljs.registerLanguage('html', xml);
hljs.registerLanguage('django', django);
hljs.registerLanguage('go', go);
hljs.registerLanguage('shell', shell);
hljs.registerLanguage('javascript', javascript);
document.addEventListener('DOMContentLoaded', () => {
  hljs.highlightAll();
});
