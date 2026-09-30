// Build a file → section → builtin outline from the generated Rye docs.
// IDs use ordinals, not names: sections and builtins can share the same name.
(function () {
    'use strict';

    // Show kind//Method as a kind badge followed by the method name, both in
    // the reference heading and in the navigation link. Keep the original
    // text in the DOM for search and accessible names.
    function formatMethod(node, name) {
        const separator = name.indexOf('//');
        if (separator < 1 || separator + 2 >= name.length) {
            node.textContent = name;
            return;
        }
        const tag = document.createElement('span');
        tag.className = 'menu-kind-tag';
        tag.textContent = name.slice(0, separator);
        node.replaceChildren(tag, document.createTextNode(name.slice(separator + 2)));
        node.setAttribute('aria-label', name);
    }

    function linkTo(heading, id) {
        heading.id = id;
        const link = document.createElement('a');
        link.href = '#' + id;
        const name = heading.textContent.trim();
        if (heading.tagName === 'H3') {
            formatMethod(heading, name);
            formatMethod(link, name);
        } else {
            link.textContent = name;
        }
        return link;
    }

    function itemFor(heading, id) {
        const li = document.createElement('li');
        li.appendChild(linkTo(heading, id));
        return li;
    }

    function buildOutline() {
        const nav = document.getElementById('outline');
        const tree = document.createElement('ul');
        tree.className = 'file-list';
        let fileIndex = 0;
        for (const file of document.querySelectorAll('main.content > h2')) {
            const fileItem = itemFor(file, 'file-' + ++fileIndex);
            if (file.title) fileItem.firstElementChild.title = file.title;
            const sections = document.createElement('ul');
            sections.className = 'section-list';
            // regen wraps all sections from a Go source file in one div.
            const body = file.nextElementSibling?.nextElementSibling;
            if (body?.classList.contains('section')) {
                let sectionIndex = 0;
                for (const section of body.querySelectorAll(':scope > h2')) {
                    const sectionItem = itemFor(section, `file-${fileIndex}-section-${++sectionIndex}`);
                    const words = document.createElement('ul');
                    words.className = 'builtin-list';
                    const sectionBody = section.nextElementSibling?.nextElementSibling;
                    if (sectionBody?.classList.contains('section')) {
                        let wordIndex = 0;
                        for (const word of sectionBody.querySelectorAll(':scope > h3')) {
                            words.appendChild(itemFor(word, `file-${fileIndex}-section-${sectionIndex}-word-${++wordIndex}`));
                        }
                    }
                    if (words.children.length) sectionItem.appendChild(words);
                    sections.appendChild(sectionItem);
                }
            }
            if (sections.children.length) fileItem.appendChild(sections);
            tree.appendChild(fileItem);
        }
        nav.appendChild(tree);

        const search = document.getElementById('menu-search');
        search.addEventListener('input', function () {
            const query = search.value.trim().toLocaleLowerCase();
            for (const li of tree.querySelectorAll('li')) {
                li.hidden = false;
            }
            if (!query) return;
            // Match at each level. A matching file/section reveals its entire
            // subtree; a matching builtin keeps its ancestors visible.
            function filter(li, inheritedMatch) {
                const selfMatches = li.firstElementChild.textContent.toLocaleLowerCase().includes(query);
                const children = Array.from(li.querySelectorAll(':scope > ul > li'));
                const childMatches = children.map(child => filter(child, inheritedMatch || selfMatches)).some(Boolean);
                li.hidden = !(inheritedMatch || selfMatches || childMatches);
                return !li.hidden;
            }
            for (const li of tree.children) filter(li, false);
        });

        const current = (location.pathname.split('/').pop() || 'base.html').replace(/\.html$/, '');
        document.getElementById('maintab-' + current)?.setAttribute('aria-current', 'page');
    }

    document.addEventListener('DOMContentLoaded', buildOutline);
})();
