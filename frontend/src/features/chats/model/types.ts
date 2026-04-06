export interface Chat {
  id: string;
  deal_id?: string;
  participants: string[];
  created_at: string;
}

export interface Message {
  id: string;
  chat_id: string;
  sender_id: string;
  content: string;
  created_at: string;
}

export interface UserInfo {
  id: string;
  name: string;
}

export interface CreateChatRequest {
  participant_id: string;
}

export interface SendMessageRequest {
  content: string;
}
