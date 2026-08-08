import { LoginForm } from '../components/LoginForm';

type LoginPageProps = {
  onSuccess: () => void;
  onGoRegister: () => void;
};

export function LoginPage({ onSuccess, onGoRegister }: LoginPageProps) {
  return <LoginForm onSuccess={onSuccess} onGoRegister={onGoRegister} />;
}
