export function buildFilter(items: string[]): any {
    const a: any = {
        $in: items.filter((a) => !a.startsWith("-")),
        $nin: items.filter((a) => a.startsWith("-")).map((a) => a.substring(1)),
    };
    if (a.$in.length === 0) {
        delete a.$in;
    }
    if (a.$nin.length === 0) {
        delete a.$nin;
    }
    return a;
}
