---
name: ring:md-to-html
description: Transform a markdown file into a standalone, styled HTML page
argument-hint: "[markdown-file]"
---
Load the visual-explainer skill, then read the markdown file provided as an argument and generate a comprehensive visual HTML document as a self-contained page.

Follow the visual-explainer skill workflow. Read the reference templates and CSS patterns before generating. This is a text-heavy markdown conversion, so you should prefer an editorial, paper/ink, or blueprint aesthetic with clear typography and hierarchy.

**Inputs:**
- Target file: `$1` (path to a markdown file)

**Data gathering phase** — read and process the markdown:
1. **Read the file in full.** Extract the structure, headings, paragraphs, lists, and code blocks.
2. **Identify visual opportunities.** Look for data that could be better represented. If you see a complex markdown table (4+ rows or 3+ columns), transform it into a styled HTML table using the data-table pattern. If you see text describing a process, sequence, or architecture, consider inserting a Mermaid diagram to supplement the text.

**Diagram structure** — the page should include:
1. **Document Header** — the title and subtitle of the document. Use a large, bold display font.
2. **Table of Contents (Optional)** — if the document is long (3+ sections), include the responsive sticky sidebar TOC pattern from `css-patterns.md`.
3. **Content Translation** — translate the markdown into semantic HTML:
   - Headings (`<h1>` to `<h4>`)
   - Paragraphs (`<p>`)
   - Lists (`<ul>`, `<ol>`, `<li>`) with proper spacing
   - Code blocks (`<pre><code>`) with subtle background and monospaced font
   - Inline code (`<code>`)
   - Blockquotes (`<blockquote>` or styled callouts)
4. **Enhanced Visuals (Proactive)**:
   - Convert ascii/markdown tables into the `data-table` layout pattern.
   - If a section describes a flow or architecture, generate a corresponding Mermaid diagram and insert it using the `.mermaid-wrap` container with zoom controls.

**Visual hierarchy**: Use the Lerian palette defined in the CSS patterns. Ensure code blocks are recessed (using the `.node--recessed` pattern if applicable) and callouts use the `.callout` pattern.

Include responsive section navigation if appropriate. Write the final self-contained HTML to `~/.agent/diagrams/` and open it in the browser.

Ultrathink.

$@