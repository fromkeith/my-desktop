export * from "./server";
import type { IPersonInfo } from "./server";

export enum WindowType {
    EmailList = "EmailListWindow",
    EmailContents = "EmailContentsWindow",
    ComposeEmail = "ComposeEmailWindow",
    LoginEmail = "LoginWindow",
    ContactList = "ContactListWindow",
}

export enum ComposeType {
    Forward = "Forward",
    Reply = "Reply",
    ReplyAll = "ReplyAll",
    New = "New",
}

export interface IComposeEmailMeta {
    to: IPersonInfo[];
    cc?: IPersonInfo[];
    bcc?: IPersonInfo[];
    subject?: string;
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
    from?: string;
}

export interface IAuthToken {
    iss: string; // issuer
    sub: string; // subject (aka account id)
    exp: number; // expires in unix timestamp
    nbf: number; // not valid before
    iat: number; // expires at
}
