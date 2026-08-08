import {
  TotalAverage,
  TotalBlock,
  TotalLabel,
  TotalsGrid,
  TotalValue,
} from '../styled-components/analytics.styles';
import { computePeriodAverages, formatNumber } from '../utils/format';
import type { StatsSummary } from '../utils/types';
import { useI18n } from '../../../i18n/I18nProvider';

type Props = {
  summary: StatsSummary;
};

export function TotalsRow({ summary }: Props) {
  const { t } = useI18n();
  const averages = computePeriodAverages(summary);

  return (
    <TotalsGrid>
      <TotalBlock>
        <TotalLabel>{t('calories')}</TotalLabel>
        <TotalValue>{formatNumber(summary.total_calories)}</TotalValue>
        <TotalAverage>
          {t('avgPrefix')} {formatNumber(averages.calories)} {t('perDay')} · {averages.days}{' '}
          {t('loggedDays')}
        </TotalAverage>
      </TotalBlock>
      <TotalBlock>
        <TotalLabel>{t('protein')}</TotalLabel>
        <TotalValue>
          {formatNumber(summary.total_protein, 1)} {t('gramsShort')}
        </TotalValue>
        <TotalAverage>
          {t('avgPrefix')} {formatNumber(averages.protein, 1)} {t('gramsShort')} {t('perDay')}
        </TotalAverage>
      </TotalBlock>
      <TotalBlock>
        <TotalLabel>{t('fat')}</TotalLabel>
        <TotalValue>
          {formatNumber(summary.total_fat, 1)} {t('gramsShort')}
        </TotalValue>
        <TotalAverage>
          {t('avgPrefix')} {formatNumber(averages.fat, 1)} {t('gramsShort')} {t('perDay')}
        </TotalAverage>
      </TotalBlock>
      <TotalBlock>
        <TotalLabel>{t('carbs')}</TotalLabel>
        <TotalValue>
          {formatNumber(summary.total_carbs, 1)} {t('gramsShort')}
        </TotalValue>
        <TotalAverage>
          {t('avgPrefix')} {formatNumber(averages.carbs, 1)} {t('gramsShort')} {t('perDay')}
        </TotalAverage>
      </TotalBlock>
    </TotalsGrid>
  );
}
