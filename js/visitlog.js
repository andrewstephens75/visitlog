function visitlog() {
    if (document.readyState !== "complete") {
        return;
    }
    var data = new FormData();
    data.append("uri", window.location.pathname);

    navigator.sendBeacon("/visitlog/log", data);
}

if (document.readyState === "complete") {
    visitlog();
} else {
    document.addEventListener('readystatechange', visitlog, false);
}
 