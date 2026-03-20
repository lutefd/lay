export interface ChatMessage {
  role: 'user' | 'assistant';
  content: string;
  images?: string[]; // base64-encoded image data (no prefix)
}
