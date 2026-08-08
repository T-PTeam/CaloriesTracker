import styled from 'styled-components';
import { theme } from '../features/analytics/styled-components/analytics.styles';
import type { Locale } from './messages';
import { useI18n } from './I18nProvider';

const Wrap = styled.div`
  display: inline-flex;
  gap: 0.35rem;
  align-items: center;
`;

const LangButton = styled.button<{ $active?: boolean }>`
  border: 1px solid ${({ $active }) => ($active ? theme.colors.accent : theme.colors.line)};
  background: ${({ $active }) => ($active ? theme.colors.accentSoft : 'transparent')};
  color: ${theme.colors.ink};
  border-radius: 0.45rem;
  padding: 0.35rem 0.65rem;
  cursor: pointer;
  font-size: 0.85rem;

  &:hover {
    border-color: ${theme.colors.accent};
  }
`;

const Label = styled.span`
  color: ${theme.colors.muted};
  font-size: 0.85rem;
  margin-right: 0.25rem;
`;

const OPTIONS: { value: Locale; label: string }[] = [
  { value: 'uk', label: 'UA' },
  { value: 'en', label: 'EN' },
];

type Props = {
  showLabel?: boolean;
};

export function LanguageSwitcher({ showLabel = false }: Props) {
  const { locale, setLocale, t } = useI18n();

  return (
    <Wrap>
      {showLabel ? <Label>{t('language')}</Label> : null}
      {OPTIONS.map((option) => (
        <LangButton
          key={option.value}
          type="button"
          $active={locale === option.value}
          onClick={() => setLocale(option.value)}
        >
          {option.label}
        </LangButton>
      ))}
    </Wrap>
  );
}
