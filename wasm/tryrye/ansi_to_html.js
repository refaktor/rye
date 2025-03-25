/**
 * Convert ANSI escape sequences to HTML
 * Based on ansi-to-html library
 */

(function() {
  'use strict';

  // ANSI color codes to CSS colors
  const foregroundColors = {
    '30': 'black',
    '31': 'red',
    '32': 'green',
    '33': 'yellow',
    '34': 'blue',
    '35': 'magenta',
    '36': 'cyan',
    '37': 'white',
    '90': '#888', // bright black (gray)
    '91': '#f55', // bright red
    '92': '#5f5', // bright green
    '93': '#ff5', // bright yellow
    '94': '#55f', // bright blue
    '95': '#f5f', // bright magenta
    '96': '#5ff', // bright cyan
    '97': '#fff'  // bright white
  };

  const backgroundColors = {
    '40': 'black',
    '41': 'red',
    '42': 'green',
    '43': 'yellow',
    '44': 'blue',
    '45': 'magenta',
    '46': 'cyan',
    '47': 'white',
    '100': '#888', // bright black (gray)
    '101': '#f55', // bright red
    '102': '#5f5', // bright green
    '103': '#ff5', // bright yellow
    '104': '#55f', // bright blue
    '105': '#f5f', // bright magenta
    '106': '#5ff', // bright cyan
    '107': '#fff'  // bright white
  };

  // 256 color support
  function getXterm256Color(code) {
    // Basic 16 colors (0-15)
    if (code < 16) {
      const basicColors = [
        '#000', '#a00', '#0a0', '#a50', '#00a', '#a0a', '#0aa', '#aaa',
        '#555', '#f55', '#5f5', '#ff5', '#55f', '#f5f', '#5ff', '#fff'
      ];
      return basicColors[code];
    }
    
    // 216 colors (16-231): 6×6×6 cube
    if (code >= 16 && code <= 231) {
      code -= 16;
      const r = Math.floor(code / 36) * 51;
      const g = Math.floor((code % 36) / 6) * 51;
      const b = (code % 6) * 51;
      return `rgb(${r}, ${g}, ${b})`;
    }
    
    // Grayscale (232-255): 24 shades from black to white
    if (code >= 232 && code <= 255) {
      const gray = (code - 232) * 10 + 8;
      return `rgb(${gray}, ${gray}, ${gray})`;
    }
    
    return null;
  }

  // Main conversion function
  function ansiToHtml(text) {
    if (!text) return '';
    
    // Replace & with &amp; to prevent HTML injection
    text = text.replace(/&/g, '&amp;');
    
    // Replace < and > with their HTML entities
    text = text.replace(/</g, '&lt;').replace(/>/g, '&gt;');
    
    // Current style state
    let currentForeground = null;
    let currentBackground = null;
    let bold = false;
    let italic = false;
    let underline = false;
    
    // Output HTML
    let html = '';
    
    // Split the text by escape sequences
    const parts = text.split(/\x1b\[/g);
    
    // Process the first part (no escape sequence)
    html += parts[0];
    
    // Process the rest of the parts
    for (let i = 1; i < parts.length; i++) {
      const part = parts[i];
      
      // Find the end of the escape sequence
      const match = part.match(/^([0-9;]*)m([\s\S]*)/);
      
      if (!match) {
        // Not a color/style escape sequence, just append it
        html += '\x1b[' + part;
        continue;
      }
      
      const codes = match[1].split(';');
      const text = match[2];
      
      // Process the escape codes
      for (let j = 0; j < codes.length; j++) {
        const code = codes[j];
        
        if (code === '0' || code === '') {
          // Reset all styles
          currentForeground = null;
          currentBackground = null;
          bold = false;
          italic = false;
          underline = false;
        } else if (code === '1') {
          bold = true;
        } else if (code === '3') {
          italic = true;
        } else if (code === '4') {
          underline = true;
        } else if (code === '22') {
          bold = false;
        } else if (code === '23') {
          italic = false;
        } else if (code === '24') {
          underline = false;
        } else if (code === '39') {
          currentForeground = null;
        } else if (code === '49') {
          currentBackground = null;
        } else if (code >= '30' && code <= '37' || code >= '90' && code <= '97') {
          // Foreground color
          currentForeground = foregroundColors[code];
        } else if (code >= '40' && code <= '47' || code >= '100' && code <= '107') {
          // Background color
          currentBackground = backgroundColors[code];
        } else if (code === '38' || code === '48') {
          // Extended color (8-bit or 24-bit)
          const isBackground = code === '48';
          
          if (codes[j+1] === '5' && j+2 < codes.length) {
            // 8-bit color (256 colors)
            const colorCode = parseInt(codes[j+2], 10);
            const color = getXterm256Color(colorCode);
            
            if (color) {
              if (isBackground) {
                currentBackground = color;
              } else {
                currentForeground = color;
              }
            }
            
            j += 2; // Skip the next two codes
          } else if (codes[j+1] === '2' && j+4 < codes.length) {
            // 24-bit color (RGB)
            const r = parseInt(codes[j+2], 10);
            const g = parseInt(codes[j+3], 10);
            const b = parseInt(codes[j+4], 10);
            
            if (!isNaN(r) && !isNaN(g) && !isNaN(b)) {
              const color = `rgb(${r}, ${g}, ${b})`;
              
              if (isBackground) {
                currentBackground = color;
              } else {
                currentForeground = color;
              }
            }
            
            j += 4; // Skip the next four codes
          }
        }
      }
      
      // Apply the current style
      let style = '';
      
      if (currentForeground) {
        style += `color: ${currentForeground}; `;
      }
      
      if (currentBackground) {
        style += `background-color: ${currentBackground}; `;
      }
      
      if (bold) {
        style += 'font-weight: bold; ';
      }
      
      if (italic) {
        style += 'font-style: italic; ';
      }
      
      if (underline) {
        style += 'text-decoration: underline; ';
      }
      
      // Add the styled text
      if (style) {
        html += `<span style="${style}">${text}</span>`;
      } else {
        html += text;
      }
    }
    
    return html;
  }

  // Expose the function globally
  window.ansiToHtml = ansiToHtml;
})();
