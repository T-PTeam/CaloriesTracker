import { useI18n } from '../../../i18n/I18nProvider';
import {
  TotalBlock,
  TotalLabel,
  TotalsGrid,
  TotalValue,
} from '../../analytics/styled-components/analytics.styles';
import { formatNumber } from '../../analytics/utils/format';
import type { DailyTargets } from '../utils/types';

type Props = {
  targets: DailyTargets;
};

export function DailyTargetsRow({ targets }: Props) {
  const { t } = useI18n();

  return (
    <TotalsGrid>
      <TotalBlock>
        <TotalLabel>{t('dailyCalories')}</TotalLabel>
        <TotalValue>{formatNumber(targets.calories)}</TotalValue>
      </TotalBlock>
      <TotalBlock>
        <TotalLabel>{t('dailyProtein')}</TotalLabel>
        <TotalValue>
          {formatNumber(targets.protein, 1)} {t('gramsShort')}
        </TotalValue>
      </TotalBlock>
      <TotalBlock>
        <TotalLabel>{t('dailyFat')}</TotalLabel>
        <TotalValue>
          {formatNumber(targets.fat, 1)} {t('gramsShort')}
        </TotalValue>
      </TotalBlock>
      <TotalBlock>
        <TotalLabel>{t('dailyCarbs')}</TotalLabel>
        <TotalValue>
          {formatNumber(targets.carbs, 1)} {t('gramsShort')}
        </TotalValue>
      </TotalBlock>
    </TotalsGrid>
  );
}
