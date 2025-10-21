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

export interface IAuthToken {
    iss: string; // issuer
    sub: string; // subject (aka account id)
    exp: number; // expires in unix timestamp
    nbf: number; // not valid before
    iat: number; // expires at
}
