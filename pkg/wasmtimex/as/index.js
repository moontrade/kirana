let memory = new WebAssembly.Memory({initial: 1, shared: false});

async function instantiate(module, imports = {}) {
    if (!imports.env.memory) {
        imports.env.memory = memory;
    }
    const adaptedImports = {
        env: Object.assign(Object.create(globalThis), imports.env || {}, {
            "console.log"(text) {
                // ~lib/bindings/dom/console.log(~lib/string/String) => void
                text = __liftString(text >>> 0);
                console.log(text);
            },
            abort(message, fileName, lineNumber, columnNumber) {
                // ~lib/builtins/abort(~lib/string/String | null?, ~lib/string/String | null?, u32?, u32?) => void
                message = __liftString(message >>> 0);
                fileName = __liftString(fileName >>> 0);
                lineNumber = lineNumber >>> 0;
                columnNumber = columnNumber >>> 0;
                (() => {
                    // @external.js
                    throw Error(`${message} in ${fileName}:${lineNumber}:${columnNumber}`);
                })();
            },
        }),
    };

    const {exports} = await WebAssembly.instantiate(module, adaptedImports);
    // const memory = imports.env.memory;

    function __liftString(pointer) {
        if (!pointer) return null;
        if (!memory) return null;
        const
            end = pointer + new Uint32Array(memory.buffer)[pointer - 4 >>> 2] >>> 1,
            memoryU16 = new Uint16Array(memory.buffer);
        let
            start = pointer >>> 1,
            string = "";
        while (end - start > 1024) string += String.fromCharCode(...memoryU16.subarray(start, start += 1024));
        return string + String.fromCharCode(...memoryU16.subarray(start, end));
    }

    return exports;
}

async function run() {
    const {
        double,
        add
    } = await (async (url, imports) => instantiate(
        await (async () => {
            try {
                return await globalThis.WebAssembly.compileStreaming(globalThis.fetch(url));
            } catch {
                return globalThis.WebAssembly.compile(await (await import("node:fs/promises")).readFile(url));
            }
        })(), imports
    ))("./build/release.wasm", {
        env: {
            memory:  new WebAssembly.Memory({initial: 1, shared: false})
        }
    });

    document.body.innerText = `${add(1, 2)}`;
}

run();