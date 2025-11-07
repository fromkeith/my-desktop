import { type Observable } from "rxjs";
import { type Readable, writable } from "svelte/store";

/**
 * Converts an RxJS Observable to a Svelte Readable store.
 */
export function observableToStore<T>(
    observable$: Observable<T>,
    initialValue: T,
): Readable<T> {
    const { subscribe, set } = writable<T>(initialValue, (set) => {
        const subscription = observable$.subscribe({
            next: (value) => set(value),
            error: (err) => console.error("Observable error:", err),
        });

        return () => {
            subscription.unsubscribe();
        };
    });

    return {
        subscribe,
    };
}
