import {
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { ChartWrap } from '../styled-components/analytics.styles';
import { formatDayLabel } from '../utils/format';
import type { DailyStat } from '../utils/types';
import { useI18n } from '../../../i18n/I18nProvider';

type Props = {
  daily: DailyStat[];
};

export function CaloriesChart({ daily }: Props) {
  const { t, locale } = useI18n();
  const data = daily.map((item) => ({
    ...item,
    label: formatDayLabel(item.date, locale),
  }));

  return (
    <ChartWrap>
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={data}>
          <CartesianGrid stroke="rgba(232,242,236,0.12)" vertical={false} />
          <XAxis dataKey="label" stroke="#9bb5a7" tickLine={false} axisLine={false} />
          <YAxis yAxisId="kcal" stroke="#9bb5a7" tickLine={false} axisLine={false} width={48} />
          <YAxis
            yAxisId="grams"
            orientation="right"
            stroke="#9bb5a7"
            tickLine={false}
            axisLine={false}
            width={40}
          />
          <Tooltip
            contentStyle={{
              background: '#16261f',
              border: '1px solid rgba(232,242,236,0.12)',
              borderRadius: 8,
            }}
          />
          <Legend />
          <Line
            yAxisId="kcal"
            type="monotone"
            dataKey="calories"
            name={t('chartCalories')}
            stroke="#3ecf8e"
            strokeWidth={2.5}
            dot={false}
          />
          <Line
            yAxisId="grams"
            type="monotone"
            dataKey="protein"
            name={t('chartProtein')}
            stroke="#8fd3ff"
            strokeWidth={2}
            dot={false}
          />
          <Line
            yAxisId="grams"
            type="monotone"
            dataKey="fat"
            name={t('chartFat')}
            stroke="#f0b429"
            strokeWidth={2}
            dot={false}
          />
          <Line
            yAxisId="grams"
            type="monotone"
            dataKey="carbs"
            name={t('chartCarbs')}
            stroke="#c4a1ff"
            strokeWidth={2}
            dot={false}
          />
        </LineChart>
      </ResponsiveContainer>
    </ChartWrap>
  );
}
