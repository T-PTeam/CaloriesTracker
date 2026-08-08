import styled from 'styled-components';
import { theme } from '../../analytics/styled-components/analytics.styles';

export const AuthShell = styled.main`
  width: min(440px, calc(100% - 2rem));
  margin: 0 auto;
  padding: 4rem 0 3rem;
`;

export const AuthBrand = styled.p`
  margin: 0 0 0.5rem;
  font-family: ${theme.fonts.display};
  font-size: clamp(2rem, 5vw, 2.8rem);
  font-weight: 600;
  letter-spacing: -0.03em;
`;

export const AuthLead = styled.p`
  margin: 0 0 1.75rem;
  color: ${theme.colors.muted};
  line-height: 1.5;
`;

export const AuthForm = styled.form`
  display: grid;
  gap: 1rem;
`;

export const Field = styled.label`
  display: grid;
  gap: 0.4rem;
  color: ${theme.colors.muted};
  font-size: 0.92rem;
`;

export const Input = styled.input`
  width: 100%;
  border: 1px solid ${theme.colors.line};
  border-radius: 0.65rem;
  background: ${theme.colors.bgElevated};
  color: ${theme.colors.ink};
  padding: 0.75rem 0.9rem;
  outline: none;

  &:focus {
    border-color: ${theme.colors.accentSoft};
  }
`;

export const SubmitButton = styled.button`
  margin-top: 0.35rem;
  border: 0;
  border-radius: 0.65rem;
  background: ${theme.colors.accent};
  color: #062016;
  font-weight: 600;
  padding: 0.85rem 1rem;
  cursor: pointer;

  &:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
`;

export const AuthError = styled.p`
  margin: 0;
  color: ${theme.colors.danger};
  font-size: 0.92rem;
`;

export const AuthSwitch = styled.p`
  margin: 1.25rem 0 0;
  color: ${theme.colors.muted};
  font-size: 0.95rem;

  button {
    border: 0;
    background: transparent;
    color: ${theme.colors.accent};
    cursor: pointer;
    padding: 0;
  }
`;
