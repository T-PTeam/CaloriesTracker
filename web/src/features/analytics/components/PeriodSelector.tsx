import {
  DateField,
  FormInput,
  PeriodButton,
  RangeApplyButton,
  RangeFields,
  Toolbar,
} from '../styled-components/analytics.styles';
import type { PeriodKey } from '../utils/types';
import { useI18n } from '../../../i18n/I18nProvider';

type Props = {
  value: PeriodKey;
  customFrom: string;
  customTo: string;
  onPresetChange: (value: Exclude<PeriodKey, 'custom'>) => void;
  onCustomFromChange: (value: string) => void;
  onCustomToChange: (value: string) => void;
  onApplyCustom: () => void;
};

export function PeriodSelector({
  value,
  customFrom,
  customTo,
  onPresetChange,
  onCustomFromChange,
  onCustomToChange,
  onApplyCustom,
}: Props) {
  const { t } = useI18n();
  const options: { key: Exclude<PeriodKey, 'custom'>; label: string }[] = [
    { key: '7d', label: t('period7') },
    { key: '14d', label: t('period14') },
    { key: '30d', label: t('period30') },
  ];

  return (
    <Toolbar>
      {options.map((option) => (
        <PeriodButton
          key={option.key}
          type="button"
          $active={value === option.key}
          onClick={() => onPresetChange(option.key)}
        >
          {option.label}
        </PeriodButton>
      ))}
      <RangeFields>
        <DateField>
          <span>{t('dateFrom')}</span>
          <FormInput
            type="date"
            value={customFrom}
            onChange={(e) => onCustomFromChange(e.target.value)}
          />
        </DateField>
        <DateField>
          <span>{t('dateTo')}</span>
          <FormInput
            type="date"
            value={customTo}
            onChange={(e) => onCustomToChange(e.target.value)}
          />
        </DateField>
        <RangeApplyButton
          type="button"
          $active={value === 'custom'}
          onClick={onApplyCustom}
          disabled={!customFrom || !customTo || customFrom > customTo}
        >
          {t('applyRange')}
        </RangeApplyButton>
      </RangeFields>
    </Toolbar>
  );
}
