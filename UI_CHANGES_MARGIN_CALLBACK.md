# UI Changes for Margin Mode and Callback Rate

## Bot Config Form - New Fields

When creating or editing a bot configuration with **Trading Mode = Futures**, two new fields will appear:

### Location in Form

The new fields appear after the "Trading Mode & Leverage" section and before the "Amount" field.

### Visual Layout

```
┌─────────────────────────────────────────────────────────────┐
│  Trading Mode & Leverage                                    │
│  ┌──────────────────────┐  ┌──────────────────────┐       │
│  │ Trading Mode         │  │ Leverage              │       │
│  │ [Futures ▼]          │  │ [10]                 │       │
│  │ Hợp đồng tương lai   │  │ Đòn bẩy 1x-125x      │       │
│  └──────────────────────┘  └──────────────────────┘       │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│  Margin Mode & Callback Rate    ⬅️ NEW SECTION              │
│  ┌──────────────────────┐  ┌──────────────────────┐       │
│  │ Margin Mode          │  │ Callback Rate (%)    │       │
│  │ [Isolated ▼]         │  │ [1.0]                │       │
│  │ Selected: Isolated   │  │ Callback rate cho     │       │
│  │                      │  │ trailing stop (0.1-5%)│       │
│  └──────────────────────┘  └──────────────────────┘       │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│  Amount (USDT)                                              │
│  [100]                                                      │
└─────────────────────────────────────────────────────────────┘
```

## Field Details

### 1. Margin Mode (Dropdown)

- **Label**: "Margin Mode"
- **Type**: Searchable dropdown
- **Options**:
  - 🔹 **Isolated** - Ký quỹ cô lập
  - 🔹 **Crossed** - Ký quỹ chéo
- **Default**: Isolated
- **Validation**: Required when trading mode is Futures
- **Help Text**: "Selected: [chosen mode]"

### 2. Callback Rate (Number Input)

- **Label**: "Callback Rate (%)"
- **Type**: Number input
- **Range**: 0.1 - 5.0
- **Step**: 0.1
- **Default**: 1.0
- **Placeholder**: "1.0"
- **Help Text**: "Callback rate cho trailing stop (0.1-5%)"

## Conditional Display

The "Margin Mode & Callback Rate" section will:

- ✅ **Show** when Trading Mode = "Futures"
- ❌ **Hide** when Trading Mode = "Spot" or "Margin"

## Dropdown Behavior

### Margin Mode Dropdown

When you click on the Margin Mode field:

1. Shows a dropdown with 2 options
2. Can type to search/filter options
3. Each option shows:
   - Main label (e.g., "Isolated")
   - Description (e.g., "Ký quỹ cô lập")
4. On hover, option background turns light indigo
5. Clicking an option selects it and closes dropdown

```
┌──────────────────────────────────────┐
│ Margin Mode                          │
│ [Search or select...        ▼]      │
└──────────────────────────────────────┘
  ┌────────────────────────────────────┐
  │ Isolated                           │
  │ Ký quỹ cô lập                      │ ← Hover background
  ├────────────────────────────────────┤
  │ Crossed                            │
  │ Ký quỹ chéo                        │
  └────────────────────────────────────┘
```

## Form Validation

### Margin Mode:

- ✅ Must be selected when trading mode is Futures
- ✅ Must be either "ISOLATED" or "CROSSED"
- ⚠️ Backend will default to "ISOLATED" if invalid

### Callback Rate:

- ✅ Must be between 0.1 and 5.0
- ✅ Can include decimal values (e.g., 1.5)
- ⚠️ Backend will default to 1.0 if out of range

## Example Values

### Conservative Setup (Low Risk)

```
Margin Mode: Isolated
Callback Rate: 0.5%
```

- Lower callback rate = tighter trailing stop
- Isolated = risk limited to position

### Aggressive Setup (High Risk)

```
Margin Mode: Crossed
Callback Rate: 3.0%
```

- Higher callback rate = looser trailing stop
- Crossed = uses full account balance

### Balanced Setup (Recommended)

```
Margin Mode: Isolated
Callback Rate: 1.0%
```

- Default values for balanced risk
- Good starting point for most users

## Styling

All fields follow the existing design system:

- Input borders: `border border-gray-300`
- Focus state: `focus:ring-2 focus:ring-indigo-500`
- Text color: `text-gray-900`
- Help text: `text-xs text-gray-500`
- Labels: `text-sm font-medium text-gray-700`
- Dropdown hover: `hover:bg-indigo-50`

## Accessibility

- All fields have proper labels
- Input types are semantically correct
- Help text provides context
- Keyboard navigation works
- Screen reader friendly

## Responsive Behavior

- Desktop: Fields appear side-by-side (2 columns)
- Mobile/Tablet: Grid adapts to smaller screens
- Dropdowns are scrollable if needed
- Touch-friendly tap targets
