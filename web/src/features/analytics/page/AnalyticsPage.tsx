import { useCallback, useEffect, useMemo, useState, useTransition } from 'react';
import {
  createMeal,
  deleteMeal,
  fetchMeals,
  fetchStatsSummary,
  reorderMeals,
  updateMeal,
} from '../apiService/analyticsApi';
import { CaloriesChart } from '../components/CaloriesChart';
import { MealHistory } from '../components/MealHistory';
import { PeriodSelector } from '../components/PeriodSelector';
import { TotalsRow } from '../components/TotalsRow';
import {
  Brand,
  GhostButton,
  Lead,
  PageShell,
  Section,
  SectionLead,
  SectionTitle,
  StatusText,
  TopBar,
  UserMeta,
} from '../styled-components/analytics.styles';
import {
  dayKeyToRangeISO,
  fillDailySeries,
  getPeriodDayCount,
  getPeriodRange,
  shiftDayKey,
  todayKeyInKyiv,
} from '../utils/format';
import type { Meal, MealItemInput, PeriodKey, StatsSummary } from '../utils/types';
import type { PublicUser } from '../../auth/utils/types';
import { ProfileForm } from '../../profile/components/ProfileForm';
import { LanguageSwitcher } from '../../../i18n/LanguageSwitcher';
import { useI18n } from '../../../i18n/I18nProvider';

type AnalyticsPageProps = {
  user: PublicUser;
  onLogout: () => void;
};

function presetKeys(period: Exclude<PeriodKey, 'custom'>) {
  const days = getPeriodDayCount(period);
  const toKey = todayKeyInKyiv();
  const fromKey = shiftDayKey(toKey, -(days - 1));
  return { fromKey, toKey };
}

export function AnalyticsPage({ user, onLogout }: AnalyticsPageProps) {
  const { t } = useI18n();
  const [period, setPeriod] = useState<PeriodKey>('7d');
  const [customFrom, setCustomFrom] = useState(() => presetKeys('7d').fromKey);
  const [customTo, setCustomTo] = useState(() => todayKeyInKyiv());
  const [appliedCustom, setAppliedCustom] = useState<{ from: string; to: string } | null>(null);
  const [summary, setSummary] = useState<StatsSummary | null>(null);
  const [meals, setMeals] = useState<Meal[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();
  const [reloadKey, setReloadKey] = useState(0);

  const refresh = useCallback(() => setReloadKey((value) => value + 1), []);

  const rangeKeys = useMemo(() => {
    if (period === 'custom' && appliedCustom) {
      return { fromKey: appliedCustom.from, toKey: appliedCustom.to };
    }
    return presetKeys(period === 'custom' ? '7d' : period);
  }, [period, appliedCustom]);

  useEffect(() => {
    let cancelled = false;
    const { from, to } =
      period === 'custom' && appliedCustom
        ? getPeriodRange('custom', appliedCustom)
        : dayKeyToRangeISO(rangeKeys.fromKey, rangeKeys.toKey);

    setError(null);

    void (async () => {
      try {
        const [stats, mealsResponse] = await Promise.all([
          fetchStatsSummary(from, to),
          fetchMeals(from, to),
        ]);
        if (cancelled) {
          return;
        }
        startTransition(() => {
          setSummary(stats);
          setMeals(mealsResponse.meals);
        });
      } catch (err) {
        if (cancelled) {
          return;
        }
        const message = err instanceof Error ? err.message : t('analyticsLoadFailed');
        setError(message);
        if (message.includes('unauthorized')) {
          onLogout();
        }
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [period, appliedCustom, rangeKeys, onLogout, reloadKey, t]);

  function handlePresetChange(next: Exclude<PeriodKey, 'custom'>) {
    setPeriod(next);
    const keys = presetKeys(next);
    setCustomFrom(keys.fromKey);
    setCustomTo(keys.toKey);
    setAppliedCustom(null);
  }

  function handleApplyCustom() {
    if (!customFrom || !customTo || customFrom > customTo) {
      return;
    }
    setAppliedCustom({ from: customFrom, to: customTo });
    setPeriod('custom');
  }

  async function handleCreate(payload: {
    raw_text?: string;
    created_at?: string;
    items?: MealItemInput[];
  }) {
    await createMeal(payload);
    refresh();
  }

  async function handleUpdate(
    id: number,
    payload: { raw_text?: string; created_at?: string; items?: MealItemInput[] },
  ) {
    await updateMeal(id, payload);
    refresh();
  }

  async function handleDelete(id: number) {
    await deleteMeal(id);
    refresh();
  }

  async function handleReorder(dateKey: string, mealIds: number[]) {
    await reorderMeals(dateKey, mealIds);
    refresh();
  }

  const chartDaily = useMemo(() => {
    if (!summary) {
      return [];
    }
    return fillDailySeries(rangeKeys.fromKey, rangeKeys.toKey, summary.daily);
  }, [summary, rangeKeys]);

  return (
    <PageShell>
      <TopBar>
        <Brand>{t('brand')}</Brand>
        <UserMeta>
          <LanguageSwitcher showLabel />
          <span>{user.email}</span>
          <GhostButton type="button" onClick={onLogout}>
            {t('logout')}
          </GhostButton>
        </UserMeta>
      </TopBar>
      <Lead>{t('analyticsLead')}</Lead>

      <Section>
        <SectionTitle>{t('profileTitle')}</SectionTitle>
        <SectionLead>{t('profileLead')}</SectionLead>
        <ProfileForm onUnauthorized={onLogout} />
      </Section>

      <PeriodSelector
        value={period}
        customFrom={customFrom}
        customTo={customTo}
        onPresetChange={handlePresetChange}
        onCustomFromChange={setCustomFrom}
        onCustomToChange={setCustomTo}
        onApplyCustom={handleApplyCustom}
      />

      {error ? <StatusText $error>{error}</StatusText> : null}
      {isPending && !summary ? <StatusText>{t('loading')}</StatusText> : null}

      {summary ? (
        <>
          <TotalsRow summary={summary} />
          <Section>
            <SectionTitle>{t('trendTitle')}</SectionTitle>
            <SectionLead>{t('trendLead')}</SectionLead>
            <CaloriesChart daily={chartDaily} />
          </Section>
          <Section>
            <SectionTitle>{t('historyTitle')}</SectionTitle>
            <SectionLead>
              {summary.meal_count} {t('mealsInPeriod')}
            </SectionLead>
            <MealHistory
              meals={meals}
              onCreate={handleCreate}
              onUpdate={handleUpdate}
              onDelete={handleDelete}
              onReorder={handleReorder}
            />
          </Section>
        </>
      ) : null}
    </PageShell>
  );
}
