import styled, { createGlobalStyle, keyframes } from 'styled-components';

export const theme = {
  colors: {
    bg: '#0f1c17',
    bgElevated: '#16261f',
    ink: '#e8f2ec',
    muted: '#9bb5a7',
    accent: '#3ecf8e',
    accentSoft: '#1f6b4a',
    line: 'rgba(232, 242, 236, 0.12)',
    danger: '#ff7b72',
  },
  fonts: {
    display: '"Fraunces", "Iowan Old Style", Georgia, serif',
    body: '"DM Sans", "Avenir Next", "Segoe UI", sans-serif',
  },
};

const rise = keyframes`
  from {
    opacity: 0;
    transform: translateY(12px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
`;

export const GlobalStyle = createGlobalStyle`
  :root {
    color-scheme: dark;
  }

  * {
    box-sizing: border-box;
  }

  html,
  body,
  #root {
    min-height: 100%;
  }

  body {
    margin: 0;
    font-family: ${theme.fonts.body};
    color: ${theme.colors.ink};
    background:
      radial-gradient(1200px 600px at 10% -10%, #1f6b4a55, transparent 55%),
      radial-gradient(900px 500px at 90% 0%, #3ecf8e22, transparent 50%),
      linear-gradient(160deg, #0b1511 0%, #12201a 45%, #0f1c17 100%);
  }

  button,
  input,
  select {
    font: inherit;
  }
`;

export const PageShell = styled.main`
  width: min(1120px, calc(100% - 2rem));
  margin: 0 auto;
  padding: 2.5rem 0 4rem;
  animation: ${rise} 0.55s ease both;
`;

export const TopBar = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 1rem;
  margin-bottom: 0.75rem;
  flex-wrap: wrap;
`;

export const UserMeta = styled.div`
  display: flex;
  align-items: center;
  gap: 0.75rem;
  color: ${theme.colors.muted};
  font-size: 0.95rem;
`;

export const GhostButton = styled.button`
  border: 1px solid ${theme.colors.line};
  background: transparent;
  color: ${theme.colors.ink};
  border-radius: 0.55rem;
  padding: 0.45rem 0.8rem;
  cursor: pointer;

  &:hover {
    border-color: ${theme.colors.accent};
  }
`;

export const Brand = styled.p`
  margin: 0;
  font-family: ${theme.fonts.display};
  font-size: clamp(2.4rem, 6vw, 3.6rem);
  font-weight: 600;
  letter-spacing: -0.03em;
  line-height: 1.05;
`;

export const Lead = styled.p`
  margin: 0 0 2rem;
  max-width: 38rem;
  color: ${theme.colors.muted};
  font-size: 1.05rem;
  line-height: 1.55;
`;

export const Toolbar = styled.div`
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: 0.75rem;
  margin-bottom: 1.75rem;
`;

export const PeriodButton = styled.button<{ $active?: boolean }>`
  border: 1px solid ${({ $active }) => ($active ? theme.colors.accent : theme.colors.line)};
  background: ${({ $active }) => ($active ? theme.colors.accentSoft : 'transparent')};
  color: ${theme.colors.ink};
  border-radius: 0.55rem;
  padding: 0.55rem 1rem;
  cursor: pointer;
  transition:
    border-color 0.2s ease,
    background 0.2s ease,
    transform 0.2s ease;

  &:hover {
    border-color: ${theme.colors.accent};
    transform: translateY(-1px);
  }
`;

export const RangeFields = styled.div`
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: 0.65rem;
`;

export const DateField = styled.label`
  display: grid;
  gap: 0.25rem;
  color: ${theme.colors.muted};
  font-size: 0.82rem;

  input {
    min-width: 9.5rem;
  }
`;

export const RangeApplyButton = styled(PeriodButton)`
  &:disabled {
    opacity: 0.45;
    cursor: not-allowed;
    transform: none;
  }
`;
export const TotalsGrid = styled.section`
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 1rem;
  margin-bottom: 1.75rem;

  @media (max-width: 860px) {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
`;

export const TotalBlock = styled.div`
  padding: 1rem 0;
  border-top: 1px solid ${theme.colors.line};
  animation: ${rise} 0.6s ease both;

  &:nth-child(2) {
    animation-delay: 0.05s;
  }
  &:nth-child(3) {
    animation-delay: 0.1s;
  }
  &:nth-child(4) {
    animation-delay: 0.15s;
  }
`;

export const TotalLabel = styled.p`
  margin: 0 0 0.35rem;
  color: ${theme.colors.muted};
  font-size: 0.85rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
`;

export const TotalValue = styled.p`
  margin: 0;
  font-family: ${theme.fonts.display};
  font-size: 2rem;
  letter-spacing: -0.02em;
`;

export const TotalAverage = styled.p`
  margin: 0.35rem 0 0;
  color: ${theme.colors.muted};
  font-size: 0.92rem;
`;

export const Section = styled.section`
  margin-top: 2.25rem;
  animation: ${rise} 0.7s ease both;
`;

export const SectionTitle = styled.h2`
  margin: 0 0 0.35rem;
  font-family: ${theme.fonts.display};
  font-size: 1.55rem;
  font-weight: 600;
`;

export const SectionLead = styled.p`
  margin: 0 0 1.25rem;
  color: ${theme.colors.muted};
`;

export const ChartWrap = styled.div`
  width: 100%;
  height: 280px;
`;

export const MealList = styled.ul`
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 1rem;
`;

export const MealRow = styled.li`
  padding: 1rem 0;
  border-top: 1px solid ${theme.colors.line};
`;

export const MealMeta = styled.div`
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 0.45rem;
  flex-wrap: wrap;
`;

export const MealTitle = styled.p`
  margin: 0;
  font-weight: 600;
`;

export const MealMacros = styled.p`
  margin: 0;
  color: ${theme.colors.muted};
  font-size: 0.95rem;
`;

export const ItemList = styled.p`
  margin: 0.35rem 0 0;
  color: ${theme.colors.muted};
  font-size: 0.92rem;
  line-height: 1.45;
`;

export const StatusText = styled.p<{ $error?: boolean }>`
  margin: 1rem 0;
  color: ${({ $error }) => ($error ? theme.colors.danger : theme.colors.muted)};
`;

export const DayBlock = styled.section`
  margin-bottom: 2rem;
  padding-top: 0.5rem;
  border-top: 1px solid ${theme.colors.line};
`;

export const DayHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  margin: 1rem 0 0.75rem;
  flex-wrap: wrap;
`;

export const DayTitle = styled.h3`
  margin: 0;
  font-family: ${theme.fonts.display};
  font-size: 1.25rem;
  font-weight: 600;
`;

export const ActionRow = styled.div`
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 0.65rem;
`;

export const SmallButton = styled.button<{ $danger?: boolean }>`
  border: 1px solid ${({ $danger }) => ($danger ? theme.colors.danger : theme.colors.line)};
  background: transparent;
  color: ${({ $danger }) => ($danger ? theme.colors.danger : theme.colors.ink)};
  border-radius: 0.45rem;
  padding: 0.35rem 0.7rem;
  cursor: pointer;
  font-size: 0.88rem;

  &:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }

  &:hover:not(:disabled) {
    border-color: ${({ $danger }) => ($danger ? theme.colors.danger : theme.colors.accent)};
  }
`;

export const CategoryChip = styled.span`
  display: inline-block;
  margin-right: 0.35rem;
  color: ${theme.colors.accent};
  font-size: 0.82rem;
`;

export const FormPanel = styled.div`
  margin-top: 0.85rem;
  padding: 1rem 0;
  border-top: 1px solid ${theme.colors.line};
  display: grid;
  gap: 0.75rem;
`;

export const FormGrid = styled.div`
  display: grid;
  gap: 0.65rem;
`;

export const FormField = styled.label`
  display: grid;
  gap: 0.3rem;
  color: ${theme.colors.muted};
  font-size: 0.88rem;
`;

export const FormInput = styled.input`
  width: 100%;
  border: 1px solid ${theme.colors.line};
  border-radius: 0.5rem;
  background: ${theme.colors.bgElevated};
  color: ${theme.colors.ink};
  padding: 0.55rem 0.7rem;
`;

export const FormSelect = styled.select`
  width: 100%;
  border: 1px solid ${theme.colors.line};
  border-radius: 0.5rem;
  background: ${theme.colors.bgElevated};
  color: ${theme.colors.ink};
  padding: 0.55rem 0.7rem;
`;

export const FormTextarea = styled.textarea`
  width: 100%;
  min-height: 4.5rem;
  border: 1px solid ${theme.colors.line};
  border-radius: 0.5rem;
  background: ${theme.colors.bgElevated};
  color: ${theme.colors.ink};
  padding: 0.55rem 0.7rem;
  resize: vertical;
`;

export const ItemEditor = styled.div`
  display: grid;
  gap: 0.5rem;
  padding: 0.75rem 0;
  border-top: 1px solid ${theme.colors.line};
`;

export const ItemEditorGrid = styled.div`
  display: grid;
  grid-template-columns: 1.4fr repeat(5, 1fr) 1.3fr auto;
  gap: 0.4rem;

  @media (max-width: 960px) {
    grid-template-columns: 1fr 1fr;
  }
`;
