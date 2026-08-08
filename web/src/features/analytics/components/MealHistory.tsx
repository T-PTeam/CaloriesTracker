import { useState } from 'react';
import {
  ActionRow,
  DayBlock,
  DayHeader,
  DayTitle,
  ItemList,
  MealList,
  MealMacros,
  MealMeta,
  MealRow,
  MealTitle,
  SmallButton,
  StatusText,
} from '../styled-components/analytics.styles';
import { categoryLabel, formatDateTime, formatNumber, groupMealsByDay } from '../utils/format';
import type { Meal, MealItemInput } from '../utils/types';
import { MealForm } from './MealForm';
import { useI18n } from '../../../i18n/I18nProvider';

type Props = {
  meals: Meal[];
  onCreate: (payload: {
    raw_text?: string;
    created_at?: string;
    items?: MealItemInput[];
  }) => Promise<void>;
  onUpdate: (
    id: number,
    payload: { raw_text?: string; created_at?: string; items?: MealItemInput[] },
  ) => Promise<void>;
  onDelete: (id: number) => Promise<void>;
  onReorder: (dateKey: string, mealIds: number[]) => Promise<void>;
};

export function MealHistory({ meals, onCreate, onUpdate, onDelete, onReorder }: Props) {
  const { t, locale } = useI18n();
  const [editingId, setEditingId] = useState<number | null>(null);
  const [addingDate, setAddingDate] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const days = groupMealsByDay(meals, locale);

  async function run(action: () => Promise<void>) {
    setError(null);
    setBusy(true);
    try {
      await action();
      setEditingId(null);
      setAddingDate(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('operationFailed'));
    } finally {
      setBusy(false);
    }
  }

  async function moveMeal(dateKey: string, dayMeals: Meal[], index: number, direction: -1 | 1) {
    const next = index + direction;
    if (next < 0 || next >= dayMeals.length) {
      return;
    }
    const ids = dayMeals.map((meal) => meal.id);
    const swapped = [...ids];
    const tmp = swapped[index];
    swapped[index] = swapped[next];
    swapped[next] = tmp;
    await run(() => onReorder(dateKey, swapped));
  }

  if (meals.length === 0) {
    return (
      <>
        {error ? <StatusText $error>{error}</StatusText> : null}
        <StatusText>{t('noMeals')}</StatusText>
        <SmallButton
          type="button"
          disabled={busy}
          onClick={() => setAddingDate(new Date().toISOString().slice(0, 10))}
        >
          {t('addMeal')}
        </SmallButton>
        {addingDate ? (
          <MealForm
            defaultDateKey={addingDate}
            submitLabel={t('create')}
            onCancel={() => setAddingDate(null)}
            onSubmit={(payload) => run(() => onCreate(payload))}
          />
        ) : null}
      </>
    );
  }

  return (
    <>
      {error ? <StatusText $error>{error}</StatusText> : null}
      <ActionRow>
        <SmallButton
          type="button"
          disabled={busy}
          onClick={() => setAddingDate(new Date().toISOString().slice(0, 10))}
        >
          {t('addMeal')}
        </SmallButton>
      </ActionRow>

      {addingDate && !days.some((d) => d.dateKey === addingDate) ? (
        <DayBlock>
          <DayHeader>
            <DayTitle>{t('newDay')}</DayTitle>
          </DayHeader>
          <MealForm
            defaultDateKey={addingDate}
            submitLabel={t('create')}
            onCancel={() => setAddingDate(null)}
            onSubmit={(payload) => run(() => onCreate(payload))}
          />
        </DayBlock>
      ) : null}

      {days.map((day) => (
        <DayBlock key={day.dateKey}>
          <DayHeader>
            <DayTitle>{day.label}</DayTitle>
            <SmallButton
              type="button"
              disabled={busy}
              onClick={() => {
                setAddingDate(day.dateKey);
                setEditingId(null);
              }}
            >
              {t('addToDay')}
            </SmallButton>
          </DayHeader>

          {addingDate === day.dateKey ? (
            <MealForm
              defaultDateKey={day.dateKey}
              submitLabel={t('create')}
              onCancel={() => setAddingDate(null)}
              onSubmit={(payload) => run(() => onCreate(payload))}
            />
          ) : null}

          <MealList>
            {day.meals.map((meal, index) => (
              <MealRow key={meal.id}>
                <MealMeta>
                  <MealTitle>{formatDateTime(meal.created_at, locale)}</MealTitle>
                  <MealMacros>
                    {formatNumber(meal.total_calories, 0, locale)} {t('kcal')} · {t('macroP')}{' '}
                    {formatNumber(meal.total_protein, 1, locale)} · {t('macroF')}{' '}
                    {formatNumber(meal.total_fat, 1, locale)} · {t('macroC')}{' '}
                    {formatNumber(meal.total_carbs, 1, locale)}
                  </MealMacros>
                </MealMeta>
                <ItemList>
                  {meal.items
                    .map(
                      (item) =>
                        `${categoryLabel(item.category, locale)}: ${item.name} (${formatNumber(item.weight_g, 0, locale)} ${t('gramsShort')})`,
                    )
                    .join(' · ')}
                </ItemList>
                <ActionRow>
                  <SmallButton
                    type="button"
                    disabled={busy || index === 0}
                    onClick={() => void moveMeal(day.dateKey, day.meals, index, -1)}
                  >
                    {t('up')}
                  </SmallButton>
                  <SmallButton
                    type="button"
                    disabled={busy || index === day.meals.length - 1}
                    onClick={() => void moveMeal(day.dateKey, day.meals, index, 1)}
                  >
                    {t('down')}
                  </SmallButton>
                  <SmallButton
                    type="button"
                    disabled={busy}
                    onClick={() => {
                      setEditingId(meal.id);
                      setAddingDate(null);
                    }}
                  >
                    {t('edit')}
                  </SmallButton>
                  <SmallButton
                    type="button"
                    $danger
                    disabled={busy}
                    onClick={() => {
                      if (window.confirm(t('deleteConfirm'))) {
                        void run(() => onDelete(meal.id));
                      }
                    }}
                  >
                    {t('delete')}
                  </SmallButton>
                </ActionRow>
                {editingId === meal.id ? (
                  <MealForm
                    initial={meal}
                    submitLabel={t('save')}
                    onCancel={() => setEditingId(null)}
                    onSubmit={(payload) => run(() => onUpdate(meal.id, payload))}
                  />
                ) : null}
              </MealRow>
            ))}
          </MealList>
        </DayBlock>
      ))}
    </>
  );
}
