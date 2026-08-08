import { useState, type FormEvent } from 'react';
import { useI18n } from '../../../i18n/I18nProvider';
import {
  ActionRow,
  FormField,
  FormGrid,
  FormInput,
  FormPanel,
  FormSelect,
  FormTextarea,
  ItemEditor,
  ItemEditorGrid,
  SmallButton,
  StatusText,
} from '../styled-components/analytics.styles';
import { fromDatetimeLocalValue, toDatetimeLocalValue } from '../utils/format';
import { type Meal, type MealCategory, type MealItemInput } from '../utils/types';

type Mode = 'manual' | 'ai';

type MealFormProps = {
  initial?: Meal;
  defaultDateKey?: string;
  submitLabel: string;
  onCancel: () => void;
  onSubmit: (payload: {
    raw_text?: string;
    created_at?: string;
    items?: MealItemInput[];
  }) => Promise<void>;
};

const CATEGORY_OPTIONS: {
  value: MealCategory;
  labelKey: 'catProtein' | 'catSlowCarbs' | 'catLipids' | 'catFastCarbs';
}[] = [
  { value: 'high_quality_protein', labelKey: 'catProtein' },
  { value: 'long_acting_carbs', labelKey: 'catSlowCarbs' },
  { value: 'lipids', labelKey: 'catLipids' },
  { value: 'fast_acting_carbs', labelKey: 'catFastCarbs' },
];

function emptyItem(): MealItemInput {
  return {
    name: '',
    weight_g: 100,
    calories: 0,
    protein: 0,
    fat: 0,
    carbs: 0,
    category: 'high_quality_protein',
  };
}

function itemsFromMeal(meal: Meal): MealItemInput[] {
  return meal.items.map((item) => ({
    name: item.name,
    weight_g: item.weight_g,
    calories: item.calories,
    protein: item.protein,
    fat: item.fat,
    carbs: item.carbs,
    category: item.category,
  }));
}

export function MealForm({
  initial,
  defaultDateKey,
  submitLabel,
  onCancel,
  onSubmit,
}: MealFormProps) {
  const { t } = useI18n();
  const defaultDate = defaultDateKey
    ? `${defaultDateKey}T12:00`
    : toDatetimeLocalValue(new Date().toISOString());

  const [mode, setMode] = useState<Mode>(initial ? 'manual' : 'ai');
  const [createdAt, setCreatedAt] = useState(
    initial ? toDatetimeLocalValue(initial.created_at) : defaultDate,
  );
  const [rawText, setRawText] = useState(initial?.raw_text ?? '');
  const [items, setItems] = useState<MealItemInput[]>(
    initial ? itemsFromMeal(initial) : [emptyItem()],
  );
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  function updateItem(index: number, patch: Partial<MealItemInput>) {
    setItems((prev) => prev.map((item, i) => (i === index ? { ...item, ...patch } : item)));
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    setError(null);
    setLoading(true);
    try {
      if (mode === 'ai') {
        const payload: {
          raw_text: string;
          created_at?: string;
        } = {
          raw_text: rawText.trim(),
        };
        if (initial) {
          payload.created_at = fromDatetimeLocalValue(createdAt);
        }
        await onSubmit(payload);
      } else {
        const cleaned = items.filter((item) => item.name.trim());
        if (cleaned.length === 0) {
          throw new Error(t('needOneItem'));
        }
        await onSubmit({
          raw_text: rawText.trim() || undefined,
          created_at: fromDatetimeLocalValue(createdAt),
          items: cleaned,
        });
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : t('saveFailed'));
    } finally {
      setLoading(false);
    }
  }

  return (
    <FormPanel as="form" onSubmit={handleSubmit}>
      <ActionRow>
        <SmallButton type="button" onClick={() => setMode('ai')} disabled={mode === 'ai'}>
          {t('modeAi')}
        </SmallButton>
        <SmallButton type="button" onClick={() => setMode('manual')} disabled={mode === 'manual'}>
          {t('modeManual')}
        </SmallButton>
      </ActionRow>

      <FormGrid>
        <FormField>
          {t('dateTime')}
          <FormInput
            type="datetime-local"
            value={createdAt}
            onChange={(e) => setCreatedAt(e.target.value)}
            required={mode === 'manual' || Boolean(initial)}
            disabled={mode === 'ai' && !initial}
          />
          {mode === 'ai' && !initial ? <StatusText>{t('aiDateHint')}</StatusText> : null}
        </FormField>

        {mode === 'ai' ? (
          <FormField>
            {t('mealDescription')}
            <FormTextarea
              value={rawText}
              onChange={(e) => setRawText(e.target.value)}
              placeholder={t('mealPlaceholder')}
              required
            />
          </FormField>
        ) : (
          <>
            <FormField>
              {t('noteOptional')}
              <FormTextarea value={rawText} onChange={(e) => setRawText(e.target.value)} />
            </FormField>
            {items.map((item, index) => (
              <ItemEditor key={index}>
                <ItemEditorGrid>
                  <FormInput
                    placeholder={t('itemName')}
                    value={item.name}
                    onChange={(e) => updateItem(index, { name: e.target.value })}
                    required
                  />
                  <FormInput
                    type="number"
                    step="0.1"
                    placeholder={t('gramsShort')}
                    value={item.weight_g}
                    onChange={(e) => updateItem(index, { weight_g: Number(e.target.value) })}
                  />
                  <FormInput
                    type="number"
                    step="0.1"
                    placeholder={t('kcal')}
                    value={item.calories}
                    onChange={(e) => updateItem(index, { calories: Number(e.target.value) })}
                  />
                  <FormInput
                    type="number"
                    step="0.1"
                    placeholder={t('macroP')}
                    value={item.protein}
                    onChange={(e) => updateItem(index, { protein: Number(e.target.value) })}
                  />
                  <FormInput
                    type="number"
                    step="0.1"
                    placeholder={t('macroF')}
                    value={item.fat}
                    onChange={(e) => updateItem(index, { fat: Number(e.target.value) })}
                  />
                  <FormInput
                    type="number"
                    step="0.1"
                    placeholder={t('macroC')}
                    value={item.carbs}
                    onChange={(e) => updateItem(index, { carbs: Number(e.target.value) })}
                  />
                  <FormSelect
                    value={item.category}
                    onChange={(e) =>
                      updateItem(index, { category: e.target.value as MealCategory })
                    }
                  >
                    {CATEGORY_OPTIONS.map((category) => (
                      <option key={category.value} value={category.value}>
                        {t(category.labelKey)}
                      </option>
                    ))}
                  </FormSelect>
                  <SmallButton
                    type="button"
                    $danger
                    onClick={() => setItems((prev) => prev.filter((_, i) => i !== index))}
                    disabled={items.length === 1}
                  >
                    ✕
                  </SmallButton>
                </ItemEditorGrid>
              </ItemEditor>
            ))}
            <SmallButton type="button" onClick={() => setItems((prev) => [...prev, emptyItem()])}>
              {t('addItem')}
            </SmallButton>
          </>
        )}
      </FormGrid>

      {error ? <StatusText $error>{error}</StatusText> : null}

      <ActionRow>
        <SmallButton type="submit" disabled={loading}>
          {loading ? t('saving') : submitLabel}
        </SmallButton>
        <SmallButton type="button" onClick={onCancel}>
          {t('cancel')}
        </SmallButton>
      </ActionRow>
    </FormPanel>
  );
}
