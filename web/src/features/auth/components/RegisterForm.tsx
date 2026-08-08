import { useState, type FormEvent } from 'react';
import styled from 'styled-components';
import { register } from '../apiService/authApi';
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

const AuthTop = styled.div`
  display: flex;
  justify-content: flex-end;
  margin-bottom: 1rem;
`;

type RegisterFormProps = {
  onSuccess: () => void;
  onGoLogin: () => void;
};

export function RegisterForm({ onSuccess, onGoLogin }: RegisterFormProps) {
  const { t } = useI18n();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [linkCode, setLinkCode] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setLoading(true);
    try {
      await register(email, password, linkCode);
      onSuccess();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('registerFailed'));
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
      <AuthLead>{t('registerLead')}</AuthLead>
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
            autoComplete="new-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={8}
          />
        </Field>
        <Field>
          {t('telegramCode')}
          <Input
            type="text"
            inputMode="numeric"
            autoComplete="one-time-code"
            value={linkCode}
            onChange={(e) => setLinkCode(e.target.value)}
            required
            minLength={6}
            maxLength={6}
          />
        </Field>
        {error ? <AuthError>{error}</AuthError> : null}
        <SubmitButton type="submit" disabled={loading}>
          {loading ? t('registering') : t('createAccount')}
        </SubmitButton>
      </AuthForm>
      <AuthSwitch>
        {t('haveAccount')}{' '}
        <button type="button" onClick={onGoLogin}>
          {t('login')}
        </button>
      </AuthSwitch>
    </AuthShell>
  );
}
