import { WindowType, type IWindow } from "$lib/models";
import { Provider } from "svelteprovider";

class WindowProvider extends Provider<IWindow[]> {
    constructor() {
        super([]);
    }
    protected async build(): Promise<IWindow[]> {
        return [
            {
                zIndex: 0,
                windowId: "window-1234",
                props: {},
                type: WindowType.EmailList,
            },
            {
                zIndex: 1,
                windowId: "window-1235",
                props: {},
                type: WindowType.EmailList,
            },
            {
                zIndex: 2,
                windowId: "window-1236",
                props: {},
                type: WindowType.EmailContents,
            },
            {
                zIndex: 3,
                windowId: "window-1237",
                props: {},
                type: WindowType.ComposeEmail,
            },
        ];
    }
    public async moveToTop(windowId: string, curIndex: number) {
        const cur = await this.promise;
        for (const w of cur) {
            if (w.zIndex > curIndex) {
                w.zIndex--;
            }
            if (w.windowId === windowId) {
                console.log(w.zIndex);
                w.zIndex = cur.length - 1;
                console.log(w.zIndex);
            }
        }
        this.setState(Promise.resolve([...cur]));
    }
}

export const windowProvider = WindowProvider.create();
