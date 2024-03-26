fetch("counter.wasm")
    .then(function(response) {
        return response.arrayBuffer();
    })
    .then(function(buffer) {
        var dependencies = {
            "global": {},
            "env": {}
        };
        dependencies["global.Math"] = window.Math;
        var moduleBufferView = new Uint8Array(buffer);
        var myMathModule = Wasm.instantiateModule(moduleBufferView, dependencies);
        console.log(myMathModule.exports.counter);
    });
