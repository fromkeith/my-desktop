import { WindowType, type IWindow } from "$lib/models";
import { Provider } from "svelteprovider";

class WindowProvider extends Provider<Map<string, IWindow>> {
    constructor() {
        super(new Map());
    }
    protected async build(): Promise<Map<string, IWindow>> {
        return new Map([]);
    }
    public async updateWindowDim(
        windowId: string,
        x: number,
        y: number,
        width: number,
        height: number,
    ) {}
    public async moveToTop(windowId: string) {
        const cur = await this.promise;
        const curIndex = cur.get(windowId)?.zIndex ?? cur.size;
        for (const w of cur.values()) {
            if (w.zIndex > curIndex) {
                w.zIndex--;
            }
            if (w.windowId === windowId) {
                w.zIndex = cur.size - 1;
            }
        }
        this.setState(Promise.resolve(new Map(cur)));
    }
    public async close(windowId: string) {
        const cur = await this.promise;
        const curIndex = cur.get(windowId)?.zIndex ?? cur.size;
        cur.delete(windowId);
        for (const w of cur.values()) {
            if (w.zIndex > curIndex) {
                w.zIndex--;
            }
        }
        this.setState(Promise.resolve(new Map(cur)));
    }
    public async open(window: Partial<IWindow>, from?: IWindow) {
        const cur = await this.promise;
        window.zIndex = cur.size;
        window.windowId = Math.floor(Math.random() * 1000).toString();
        if (window.x === undefined) {
            window.x = (from?.x ?? 0) + 50;
        }
        if (window.y === undefined) {
            window.y = (from?.y ?? 0) + 50;
        }
        if (window.width === undefined) {
            window.width = 600;
        }
        if (window.height === undefined) {
            window.height = 800;
        }
        window.from = from?.windowId ?? undefined;
        cur.set(window.windowId, window as IWindow);
        this.setState(Promise.resolve(new Map(cur)));
    }
}

class WindowListProvider extends Provider<IWindow[]> {
    constructor() {
        super([], windowProvider());
    }

    protected async build(windows: Map<string, IWindow>): Promise<IWindow[]> {
        return Array.from(windows.values());
    }
}

export const windowProvider = WindowProvider.create();
export const windowListProvider = WindowListProvider.create();
