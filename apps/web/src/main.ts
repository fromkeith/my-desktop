import { mount } from "svelte";
import "./app.css";
import App from "./App.svelte";
import DOMPurify from "dompurify";

const app = mount(App, {
  target: document.getElementById("app")!,
});

DOMPurify.addHook("afterSanitizeAttributes", (node) => {
  if (node.tagName === "A") {
    const href = node.getAttribute("href") || "";

    // skip anchors / weird schemes
    if (!href || href.startsWith("#") || href.startsWith("javascript:")) {
      node.removeAttribute("href");
      return;
    }

    // open in new tab
    node.setAttribute("target", "_blank");

    // make sure noopener (and usually noreferrer) are present
    const rel = new Set(
      (node.getAttribute("rel") || "").split(/\s+/).filter(Boolean)
    );
    rel.add("noopener");
    node.setAttribute("rel", [...rel].join(" "));
  }
});

export default app;
