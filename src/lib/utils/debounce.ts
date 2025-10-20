export function createDebounce() {
    let t: NodeJS.Timeout;
    let abort: (() => void) | undefined;
    return (milli: number): Promise<void> => {
        clearTimeout(t);
        if (abort) {
            abort();
        }
        return new Promise<void>((resolve, reject) => {
            abort = reject;
            t = setTimeout(() => {
                abort = undefined;
                resolve();
            }, milli);
        });
    };
}
