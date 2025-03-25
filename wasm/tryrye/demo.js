$(function () {
  // Custom theme to match style of xterm.js logo
  var baseTheme = {
    foreground: '#F8F8F8',
    background: '#2D2E2C',
    selection: '#5DA5D533',
    black: '#1E1E1D',
    brightBlack: '#262625',
    red: '#CE5C5C',
    brightRed: '#FF7272',
    green: '#5BCC5B',
    brightGreen: '#72FF72',
    yellow: '#CCCC5B',
    brightYellow: '#FFFF72',
    blue: '#5D5DD3',
    brightBlue: '#7279FF',
    magenta: '#BC5ED1',
    brightMagenta: '#E572FF',
    cyan: '#5DA5D5',
    brightCyan: '#72F0FF',
    white: '#F8F8F8',
    brightWhite: '#FFFFFF'
  };
  // vscode-snazzy https://github.com/Tyriar/vscode-snazzy
  var otherTheme = {
    foreground: '#eff0eb',
    background: '#282a36',
    selection: '#97979b33',
    black: '#282a36',
    brightBlack: '#686868',
    red: '#ff5c57',
    brightRed: '#ff5c57',
    green: '#5af78e',
    brightGreen: '#5af78e',
    yellow: '#f3f99d',
    brightYellow: '#f3f99d',
    blue: '#57c7ff',
    brightBlue: '#57c7ff',
    magenta: '#ff6ac1',
    brightMagenta: '#ff6ac1',
    cyan: '#9aedfe',
    brightCyan: '#9aedfe',
    white: '#f1f1f0',
    brightWhite: '#eff0eb'
  };
  var isBaseTheme = true;

  var term = new window.Terminal({
    fontFamily: '"Cascadia Code", Menlo, monospace',
    theme: baseTheme,
    cursorBlink: true,
    allowProposedApi: true
  });
  term.open(document.querySelector('.demo .inner'));

  var isWebglEnabled = false;
  try {
    const webgl = new window.WebglAddon.WebglAddon();
    term.loadAddon(webgl);
    isWebglEnabled = true;
  } catch (e) {
    console.warn('WebGL addon threw an exception during load', e);
  }

  // Cancel wheel events from scrolling the page if the terminal has scrollback
  document.querySelector('.xterm').addEventListener('wheel', e => {
    if (term.buffer.active.baseY > 0) {
      e.preventDefault();
    }
  });

  function runFakeTerminal() {
    if (term._initialized) {
      return;
    }

    term._initialized = true;

    term.prompt = () => {
      term.write('\r\n$ ');
    };

    // TODO: Use a nicer default font
    term.writeln([
      '    Xterm.js is the frontend component that powers many terminals including',
      '                           \x1b[3mVS Code\x1b[0m, \x1b[3mHyper\x1b[0m and \x1b[3mTheia\x1b[0m!',
      '',
      ' ┌ \x1b[1mFeatures\x1b[0m ──────────────────────────────────────────────────────────────────┐',
      ' │                                                                            │',
      ' │  \x1b[31;1mApps just work                         \x1b[32mPerformance\x1b[0m                        │',
      ' │   Xterm.js works with most terminal      Xterm.js is fast and includes an  │',
      ' │   apps like bash, vim and tmux           optional \x1b[3mWebGL renderer\x1b[0m           │',
      ' │                                                                            │',
      ' │  \x1b[33;1mAccessible                             \x1b[34mSelf-contained\x1b[0m                     │',
      ' │   A screen reader mode is available      Zero external dependencies        │',
      ' │                                                                            │',
      ' │  \x1b[35;1mUnicode support                        \x1b[36mAnd much more...\x1b[0m                   │',
      ' │   Supports CJK 語 and emoji \u2764\ufe0f            \x1b[3mLinks\x1b[0m, \x1b[3mthemes\x1b[0m, \x1b[3maddons\x1b[0m,            │',
      ' │                                          \x1b[3mtyped API\x1b[0m, \x1b[3mdecorations\x1b[0m            │',
      ' │                                                                            │',
      ' └────────────────────────────────────────────────────────────────────────────┘',
      ''
    ].join('\n\r'));

    term.writeln('Below is a simple emulated backend, try running `help`.');
    addDecoration(term);
    prompt(term);

    term.onData(e => {
      switch (e) {
        case '\u0003': // Ctrl+C
          term.write('^C');
          prompt(term);
          break;
        case '\r': // Enter
          runCommand(term, command);
          command = '';
          break;
        case '\u007F': // Backspace (DEL)
          // Do not delete the prompt
          if (term._core.buffer.x > 2) {
            term.write('\b \b');
            if (command.length > 0) {
              command = command.substr(0, command.length - 1);
            }
          }
          break;
        default: // Print all other characters for demo
          if (e >= String.fromCharCode(0x20) && e <= String.fromCharCode(0x7E) || e >= '\u00a0') {
            command += e;
            term.write(e);
          }
      }
    });

    // Create a very simple link provider which hardcodes links for certain lines
    term.registerLinkProvider({
      provideLinks(bufferLineNumber, callback) {
        switch (bufferLineNumber) {
          case 2:
            callback([
              {
                text: 'VS Code',
                range: { start: { x: 28, y: 2 }, end: { x: 34, y: 2 } },
                activate() {
                  window.open('https://github.com/microsoft/vscode', '_blank');
                }
              },
              {
                text: 'Hyper',
                range: { start: { x: 37, y: 2 }, end: { x: 41, y: 2 } },
                activate() {
                  window.open('https://github.com/vercel/hyper', '_blank');
                }
              },
              {
                text: 'Theia',
                range: { start: { x: 47, y: 2 }, end: { x: 51, y: 2 } },
                activate() {
                  window.open('https://github.com/eclipse-theia/theia', '_blank');
                }
              }
            ]);
            return;
          case 8:
            callback([
              {
                text: 'WebGL renderer',
                range: { start: { x: 54, y: 8 }, end: { x: 67, y: 8 } },
                activate() {
                  window.open('https://npmjs.com/package/xterm-addon-webgl', '_blank');
                }
              }
            ]);
            return;
          case 14:
            callback([
              {
                text: 'Links',
                range: { start: { x: 45, y: 14 }, end: { x: 49, y: 14 } },
                activate() {
                  window.alert('You can handle links any way you want');
                }
              },
              {
                text: 'themes',
                range: { start: { x: 52, y: 14 }, end: { x: 57, y: 14 } },
                activate() {
                  isBaseTheme = !isBaseTheme;
                  term.options.theme = isBaseTheme ? baseTheme : otherTheme;
                  document.querySelector('.demo .inner').classList.toggle('other-theme', !isBaseTheme);
                  term.write(`\r\nActivated ${isBaseTheme ? 'xterm.js' : 'snazzy'} theme`);
                  prompt(term);
                }
              },
              {
                text: 'addons',
                range: { start: { x: 60, y: 14 }, end: { x: 65, y: 14 } },
                activate() {
                  window.open('/docs/guides/using-addons/', '_blank');
                }
              }
            ]);
            return;
          case 15: callback([
            {
              text: 'typed API',
              range: { start: { x: 45, y: 15 }, end: { x: 53, y: 15 } },
              activate() {
                window.open('https://github.com/xtermjs/xterm.js/blob/master/typings/xterm.d.ts', '_blank');
              }
            },
            {
              text: 'decorations',
              range: { start: { x: 56, y: 15 }, end: { x: 66, y: 15 } },
              activate() {
                window.open('https://github.com/xtermjs/xterm.js/blob/a351f5758a5126308b90d60b604b528462f6f051/typings/xterm.d.ts#L372', '_blank');
              }
            },
          ]);
            return;
        }
        callback(undefined);
      }
    });
  }

  function prompt(term) {
    command = '';
    term.write('\r\n$ ');
  }

  var command = '';
  var commands = {
    help: {
      f: () => {
        const padding = 10;
        function formatMessage(name, description) {
          const maxLength = term.cols - padding - 3;
          let remaining = description;
          const d = [];
          while (remaining.length > 0) {
            // Trim any spaces left over from the previous line
            remaining = remaining.trimStart();
            // Check if the remaining text fits
            if (remaining.length < maxLength) {
              d.push(remaining);
              remaining = '';
            } else {
              let splitIndex = -1;
              // Check if the remaining line wraps already
              if (remaining[maxLength] === ' ') {
                splitIndex = maxLength;
              } else {
                // Find the last space to use as the split index
                for (let i = maxLength - 1; i >= 0; i--) {
                  if (remaining[i] === ' ') {
                    splitIndex = i;
                    break;
                  }
                }
              }
              d.push(remaining.substring(0, splitIndex));
              remaining = remaining.substring(splitIndex);
            }
          }
          const message = (
            `  \x1b[36;1m${name.padEnd(padding)}\x1b[0m ${d[0]}` +
            d.slice(1).map(e => `\r\n  ${' '.repeat(padding)} ${e}`)
          );
          return message;
        }
        term.writeln([
          'Welcome to xterm.js! Try some of the commands below.',
          '',
          ...Object.keys(commands).map(e => formatMessage(e, commands[e].description))
        ].join('\n\r'));
        prompt(term);
      },
      description: 'Prints this help message',
    },
    ls: {
      f: () => {
        term.writeln(['a', 'bunch', 'of', 'fake', 'files'].join('\r\n'));
        term.prompt(term);
      },
      description: 'Prints a fake directory structure'
    },
    loadtest: {
      f: () => {
        let testData = [];
        let byteCount = 0;
        for (let i = 0; i < 50; i++) {
          let count = 1 + Math.floor(Math.random() * 79);
          byteCount += count + 2;
          let data = new Uint8Array(count + 2);
          data[0] = 0x0A; // \n
          for (let i = 1; i < count + 1; i++) {
            data[i] = 0x61 + Math.floor(Math.random() * (0x7A - 0x61));
          }
          // End each line with \r so the cursor remains constant, this is what ls/tree do and improves
          // performance significantly due to the cursor DOM element not needing to change
          data[data.length - 1] = 0x0D; // \r
          testData.push(data);
        }
        let start = performance.now();
        for (let i = 0; i < 1024; i++) {
          for (const d of testData) {
            term.write(d);
          }
        }
        // Wait for all data to be parsed before evaluating time
        term.write('', () => {
          let time = Math.round(performance.now() - start);
          let mbs = ((byteCount / 1024) * (1 / (time / 1000))).toFixed(2);
          term.write(`\n\r\nWrote ${byteCount}kB in ${time}ms (${mbs}MB/s) using the ${isWebglEnabled ? 'webgl' : 'canvas'} renderer`);
          term.prompt();
        });
      },
      description: 'Simulate a lot of data coming from a process'
    },
    chars: {
      f: () => {
        const _1to8 = [];
        for (let i = 1; i <= 8; i++) {
          _1to8.push(i);
        }
        const _1to16 = [];
        for (let i = 1; i <= 16; i++) {
          _1to16.push(i);
        }
        const _1to24 = [];
        for (let i = 1; i <= 24; i++) {
          _1to24.push(i);
        }
        const _1to32 = [];
        for (let i = 1; i <= 32; i++) {
          _1to32.push(i);
        }
        const _0to35 = [];
        for (let i = 0; i <= 35; i++) {
          _0to35.push(i);
        }
        const _1to64 = [];
        for (let i = 1; i <= 64; i++) {
          _1to64.push(i);
        }
        const _0to255 = [];
        for (let i = 17; i <= 255; i++) {
          _0to255.push(i);
        }
        const lines = [
          ['Ascii ─', 'abc123'],
          ['CJK ─', '汉语, 漢語, 日本語, 한국어'],
          ['Powerline ─', '\ue0b2\ue0b0\ue0b3\ue0b1\ue0b6\ue0b4\ue0b7\ue0b5\ue0ba\ue0b8\ue0bd\ue0b9\ue0be\ue0bc'],
          ['Box drawing ┬', '┌─┬─┐ ┏━┳━┓ ╔═╦═╗ ┌─┲━┓ ╲   ╱'],
          ['            │', '│ │ │ ┃ ┃ ┃ ║ ║ ║ │ ┃ ┃  ╲ ╱'],
          ['            │', '├─┼─┤ ┣━╋━┫ ╠═╬═╣ ├─╄━┩   ╳'],
          ['            │', '│ │ │ ┃ ┃ ┃ ║ ║ ║ │ │ │  ╱ ╲'],
          ['            └', '└─┴─┘ ┗━┻━┛ ╚═╩═╝ └─┴─┘ ╱   ╲'],
          ['Block elem ─', '░▒▓█ ▁▂▃▄▅▆▇█ ▏▎▍▌▋▊▉'],
          ['Emoji ─', '😉 👋'],
          ['16 color ─', [..._1to8.map(e => `\x1b[3${e - 1}m●`), ..._1to8.map(e => `\x1b[1;3${e - 1}m●`)].join('')],
          ['256 color ┬', [..._0to35.map(e => `\x1b[38;5;${16 + 36 * 0 + e}m●`)].join('')],
          ['          │', [..._0to35.map(e => `\x1b[38;5;${16 + 36 * 1 + e}m●`)].join('')],
          ['          │', [..._0to35.map(e => `\x1b[38;5;${16 + 36 * 2 + e}m●`)].join('')],
          ['          │', [..._0to35.map(e => `\x1b[38;5;${16 + 36 * 3 + e}m●`)].join('')],
          ['          │', [..._0to35.map(e => `\x1b[38;5;${16 + 36 * 4 + e}m●`)].join('')],
          ['          │', [..._0to35.map(e => `\x1b[38;5;${16 + 36 * 5 + e}m●`)].join('')],
          ['          └', [..._1to24.map(e => `\x1b[38;5;${232 + e - 1}m●`)].join('')],
          ['True color ┬', [..._1to64.map(e => `\x1b[38;2;${64 * 0 + e - 1};0;0m●`)].join('')],
          ['           │', [..._1to64.map(e => `\x1b[38;2;${64 * 1 + e - 1};0;0m●`)].join('')],
          ['           │', [..._1to64.map(e => `\x1b[38;2;${64 * 2 + e - 1};0;0m●`)].join('')],
          ['           └', [..._1to64.map(e => `\x1b[38;2;${64 * 3 + e - 1};0;0m●`)].join('')],
          ['Styles ─', ['\x1b[1mBold', '\x1b[2mFaint', '\x1b[3mItalics', '\x1b[7mInverse', '\x1b[9mStrikethrough', '\x1b[8mInvisible'].join('\x1b[0m, ')],
          ['Underlines ─', ['\x1b[4:1mStraight', '\x1b[4:2mDouble', '\x1b[4:3mCurly', '\x1b[4:4mDotted', '\x1b[4:5mDashed'].join('\x1b[0m, ')],
        ];
        const maxLength = lines.reduce((p, c) => Math.max(p, c[0].length), 0);
        term.write('\r\n');
        term.writeln(lines.map(e => `${e[0].padStart(maxLength)}  ${e[1]}\x1b[0m`).join('\r\n'));
        term.prompt();
      },
      description: 'Prints a wide range of characters and styles that xterm.js can handle'
    }
  };

  function runCommand(term, text) {
    const command = text.trim().split(' ')[0];
    if (command.length > 0) {
      term.writeln('');
      if (command in commands) {
        commands[command].f();
        return;
      }
      term.writeln(`${command}: command not found`);
    }
    prompt(term);
  }

  runFakeTerminal();
});

function addDecoration(term) {
  const marker = term.registerMarker(15);
  const decoration = term.registerDecoration({ marker, x: 44 });
  decoration.onRender(element => {
    element.classList.add('link-hint-decoration');
    element.innerText = 'Try clicking italic text';
    // must be inlined to override inlined width/height coming from xterm
    element.style.height = '';
    element.style.width = '';
  });
}
