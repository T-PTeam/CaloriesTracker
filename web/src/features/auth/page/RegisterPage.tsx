import { RegisterForm } from '../components/RegisterForm';

type RegisterPageProps = {
  onSuccess: () => void;
  onGoLogin: () => void;
};

export function RegisterPage({ onSuccess, onGoLogin }: RegisterPageProps) {
  return <RegisterForm onSuccess={onSuccess} onGoLogin={onGoLogin} />;
}
