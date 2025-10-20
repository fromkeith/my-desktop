export enum WindowType {
    EmailList = "EmailListWindow",
    EmailContents = "EmailContentsWindow",
    ComposeEmail = "ComposeEmailWindow",
}

export interface IWindow {
    zIndex: number;
    windowId: string;
    props: Object;
    type: WindowType;
    x: number;
    y: number;
    width: number;
    height: number;
}
