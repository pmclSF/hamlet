export interface ModalProps {
  title: string;
  content: string;
  isOpen: boolean;
  onClose: () => void;
}

export function Modal({ title, content, isOpen, onClose }: ModalProps) {
  if (!isOpen) return null;
  return `<div class="modal"><h2>${title}</h2><p>${content}</p><button onclick="${onClose}">Close</button></div>`;
}
