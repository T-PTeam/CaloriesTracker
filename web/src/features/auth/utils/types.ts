export type PublicUser = {
  id: number;
  email: string;
  telegram_id: number;
  language?: 'uk' | 'en';
};

export type AuthResult = {
  token: string;
  user: PublicUser;
};
