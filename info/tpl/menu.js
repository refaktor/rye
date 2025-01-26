function escapeHtml(text) {
    const map = {
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;',
        '"': '&quot;',
        "'": '&#039;'
    };
    return text.replace(/[&<>"']/g, function(m) { return map[m]; });
}

function generateMenuFromHeadings(node, hh) {
    // Select all H2 elements
    const h2Elements = node.querySelectorAll(hh);
    
    // Create a menu container, for example, a <ul> element
    const menu = document.createElement('ul');

    // Iterate over each H2 element
    h2Elements.forEach((h2, index) => {
        // Create a menu item, for example, an <li> element
        const menuItem = document.createElement('li');

        // Set the text of the menu item to the text of the H2 element
        // menuItem.textContent = h2.textContent;

        // Optionally, set an id on the H2 for navigation
	var index = escapeHtml(h2.textContent);
	
        const h2Id = `heading-${index}`;
        h2.setAttribute('id', h2Id);

        // Optionally, create a link for navigation
        const link = document.createElement('a');
        link.setAttribute('href', `#${h2Id}`);
        link.textContent = h2.textContent;
        menuItem.appendChild(link);

	var div = _dom.seekFwd(h2, "DIV");

	if (div.className == "section") {
	    var submenu = generateMenuFromHeadings(div, "h3");
	    if (submenu) {
		menuItem.appendChild(submenu);
	    }
	}
	
        // Append the menu item to the menu
        menu.appendChild(menuItem);
    });

    // Append the menu to the document, for example, to the body or a specific div
    return menu;
    // document.getElementById("menu-holder").appendChild(menu);
}

function generateMenu() {
    var menu = generateMenuFromHeadings(document, "h2");
    document.getElementById("menu-holder").appendChild(menu);
}

function generateMenuFromH2_original(div) {
    // Select all H2 elements
    const h2Elements = document.querySelectorAll('h2');

    // Create a menu container, for example, a <ul> element
    const menu = document.createElement('ul');

    // Iterate over each H2 element
    h2Elements.forEach((h2, index) => {
        // Create a menu item, for example, an <li> element
        const menuItem = document.createElement('li');

        // Set the text of the menu item to the text of the H2 element
        // menuItem.textContent = h2.textContent;

        // Optionally, set an id on the H2 for navigation
        const h2Id = `heading-${index}`;
        h2.setAttribute('id', h2Id);

        // Optionally, create a link for navigation
        const link = document.createElement('a');
        link.setAttribute('href', `#${h2Id}`);
        link.textContent = h2.textContent;
        menuItem.appendChild(link);

        // Append the menu item to the menu
        menu.appendChild(menuItem);
    });

    // Append the menu to the document, for example, to the body or a specific div
    document.getElementById("menu-holder").appendChild(menu);
}


//

function styleCurrentTab() {
    var cur = document.location.pathname.match(/\/([a-z]+).html$/)[1];
    document.getElementById("maintab-"+cur).className += " current";
}
