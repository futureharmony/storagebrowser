interface PopupProps {
  prompt: string;
  confirm?: any;
  action?: PopupAction;
  saveAction?: () => void;
  props?: any;
  close?: (() => Promise<string | void> | void) | null;
  visible?: boolean;
}

type PopupAction = (e: Event) => void;
