import { derived, writable, type Writable } from "svelte/store"


const prefs = JSON.parse(localStorage.getItem("prefs") || "{}") || {}

function dump() {
    localStorage.setItem("prefs", JSON.stringify(prefs))
}

function usePref<T>(name: string, init: T): Writable<T> {
    const store = writable<T>(prefs[name] || init);
    store.subscribe((v) => {
        prefs[name] = v
        dump()
    })
    return store
}

export const mainDetailed = usePref("main.detailed", false);