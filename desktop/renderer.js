// This file is required by the index.html file and will
// be executed in the renderer process for that window.
// No Node.js APIs are available in this process because
// `nodeIntegration` is turned off. Use `preload.js` to
// selectively enable features needed in the rendering
// process.
let b = document.getElementById("clickto");
b.addEventListener("click", submit);

function submit() {
    let e = document.getElementById("in1").value,
        t = document.getElementById("in2").value,
        n = document.getElementById("in3").value;
    document.getElementById("msg");
    return 0 === e.length || e.length > 2 || "" === e ? void show("姓名不能为空或者超过2") : t > 4 ? void show("姓名长度不能超过4位") : n > 2e4 ? void show("每次生成个数不能超过20000个") : void post("/newnamedesk", {
        first: e,
        number: t,
        alln: n
    }).then(res=>res.text()).then(function (e) {
        document.getElementById("downloads").className = "btn-download", document.getElementById("download").href = e;
            document.getElementById("trueurl").value = e;
    })
}

function show(e) {
    let t = document.getElementById("msg"), n = document.createElement("div"), o = document.createElement("p");
    n.className = "model ani", o.style.textAlign = "center", o.innerText = e, n.appendChild(o), t.appendChild(n)
}

function post(e, t) {
    return fetch("https://yaop.ink/newnames" + e, {
        method: "POST", body: JSON.stringify(t), headers: {
            'Content-Type': 'application/json'
        }
    })
    // return new Promise(function (n, o) {
    //     let d = new XMLHttpRequest;
    //     d.open("POST", "https://yaop.ink/newnames" + e, !0), d.setRequestHeader("Content-Type", "application/json"), d.onreadystatechange = function () {
    //         if (4 === d.readyState) if (200 === d.status) try {
    //             n(d.responseText)
    //         } catch (e) {
    //             o(e)
    //         } else o(new Error(d.statusText))
    //     }, d.send(JSON.stringify(t))
    // })
}
