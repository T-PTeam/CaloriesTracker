import { useState, type FormEvent } from 'react';
import { login } from '../apiService/authApi';
import {
  AuthBrand,
  AuthError,
  AuthForm,
  AuthLead,
  AuthShell,
  AuthSwitch,
  Field,
  Input,
  SubmitButton,
} from '../styled-components/auth.styles';
import { LanguageSwitcher } from '../../../i18n/LanguageSwitcher';
import { useI18n } from '../../../i18n/I18nProvider';
import styled from 'styled-components';

const AuthTop = styled.div`
  display: flex;
  justify-content: flex-end;
  margin-bottom: 1rem;
`;

type LoginFormProps = {
  onSuccess: () => void;
  onGoRegister: () => void;
};

export function LoginForm({ onSuccess, onGoRegister }: LoginFormProps) {
  const { t } = useI18n();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setLoading(true);
    try {
      await login(email, password);
      onSuccess();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('loginFailed'));
    } finally {
      setLoading(false);
    }
  }

  return (
    <AuthShell>
      <AuthTop>
        <LanguageSwitcher showLabel />
      </AuthTop>
      <AuthBrand>{t('brand')}</AuthBrand>
      <AuthLead>{t('loginLead')}</AuthLead>
      <AuthForm onSubmit={handleSubmit}>
        <Field>
          {t('email')}
          <Input
            type="email"
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </Field>
        <Field>
          {t('password')}
          <Input
            type="password"
            autoComplete="current-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={8}
          />
        </Field>
        {error ? <AuthError>{error}</AuthError> : null}
        <SubmitButton type="submit" disabled={loading}>
          {loading ? t('loggingIn') : t('login')}
        </SubmitButton>
      </AuthForm>
      <AuthSwitch>
        {t('noAccount')}{' '}
        <button type="button" onClick={onGoRegister}>
          {t('registerLink')}
        </button>
      </AuthSwitch>
    </AuthShell>
  );
}
