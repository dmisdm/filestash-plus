import { createElement } from "../../lib/skeleton/index.js";
import rxjs from "../../lib/rx.js";
import { fromMarkdown } from "../../pages/viewerpage/application_form.js";

export function renderLeaf({ format, label, description, type }) {
    if (label === "banner") return createElement(`
        <div class="banner">
            ${fromMarkdown(description)}
        </div>
    `);
    const $el = createElement(`
        <label class="no-select">
            <div class="flex">
                <span class="ellipsis">
                    ${format(label)}:
                </span>
                <div style="width:100%;" data-bind="children"></div>
            </div>
        </label>
    `);
    if (type === "hidden") $el.classList.add("hidden");
    if (description) $el.appendChild(createElement(`
        <div class="flex">
            <span class="nothing"></span>
            <div style="width:100%;">
                <div class="description">${description}</div>
            </div>
        </div>
    `));
    return $el;
}

/** Flatten a backend LoginForm spec for the admin storage page (no advanced toggles). */
export function flattenBackendFields(spec) {
    if (!spec) return {};
    const flat = JSON.parse(JSON.stringify(spec));
    for (const input in flat) {
        if (flat[input]?.type === "enable") {
            delete flat[input];
        } else if (flat[input]?.id) {
            delete flat[input].id;
        }
    }
    delete flat.type;
    return flat;
}

export function connectionParamsToFormState(label, params) {
    const state = {};
    for (const [key, value] of Object.entries(params)) {
        if (key === "type" || key === "label") continue;
        if (value != null && value !== "") state[`${label}.${key}`] = value;
    }
    return state;
}

export function useForm$($inputNodeList) {
    return rxjs.pipe(
        rxjs.mergeMap(() => $inputNodeList()),
        rxjs.mergeMap(($el) => rxjs.fromEvent($el, "input")),
        rxjs.mergeMap(($el) => {
            if ($el.target.checkValidity() === false) {
                $el.target.reportValidity();
                return rxjs.EMPTY;
            }
            return rxjs.of($el);
        }),
        rxjs.mergeMap(async(e) => ({
            name: e.target.getAttribute("name"),
            value: await (async function() {
                switch (e.target.getAttribute("type")) {
                case "checkbox":
                    return e.target.checked;
                case "file":
                    if (e.target.files.length === 0) return null;
                    return await new Promise((done) => {
                        const reader = new window.FileReader();
                        reader.readAsDataURL(e.target.files[0]);
                        reader.onload = () => done(reader.result);
                    });
                default:
                    return e.target.value;
                }
            }()),
        })),
        rxjs.scan((store, keyValue) => {
            store[keyValue.name] = keyValue.value;
            return store;
        }, {})
    );
}

export function formObjToJSON$() {
    const formObjToJSON = (o, level = 0) => {
        const obj = Object.assign({}, o);
        Object.keys(obj).forEach((key) => {
            const t = obj[key];
            if ("label" in t && "type" in t && "default" in t && "value" in t) {
                let value = obj[key].value;
                if (t.type === "number") value = parseInt(value);
                obj[key] = value;
            } else {
                obj[key] = formObjToJSON(obj[key], level + 1);
            }
        });
        return obj;
    };
    return rxjs.map(formObjToJSON);
}
