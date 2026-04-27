# PoEAutoFilter

PoEAutoFilter is a powerful tool designed to automatically maintain and update your Path of Exile item filters based on real-time market data from [poe.ninja](https://poe.ninja). It ensures your filter always reflects the actual value of item stacks like Currency, Fragments, Scarabs, and more, using custom visual styles that you define.

## 🚀 Key Features

- **Real-Time Market Integration**: Fetches the latest item prices automatically.
- **Dynamic Stack Valuation**: Smartly calculates if a stack of items (e.g., 10x Glassblower's Baubles) is worth showing based on your custom thresholds.
- **Native Cross-Platform GUI**: Easy-to-use interface available for both Windows and Linux.
- **Style Library**: Create custom themes with specific colors, font sizes, alert sounds, minimap icons, and light beams.
- **Template-Based**: Merges its dynamic rules with your existing favorite filter (like a FilterBlade base), so you keep all your specialized rules for Maps, Rares, and Uniques.

## 📦 Supported Item Types

PoEAutoFilter currently manages automated valuation for:

- Currencies
- Fragments
- Scarabs
- Essences
- Fossils

## 📖 First-Time Setup

1. **Launch**: Open the `PoEAutoFilter` application.
2. **Current League**: In the **General** tab, ensure the League name matches what you are currently playing (e.g., `Standard` or the current Challenge League).
3. **Select Your Files**:
   - **Base Filter File**: Select your "source" filter (e.g., your downloaded FilterBlade file). This file is never modified.
   - **Output Filter File**: Select the `.filter` file in your Path of Exile documents folder that the game actually loads.
4. **Define Your Styles**:
   - Go to the **Style Library** and create styles (e.g., "High Value", "Mid Value"). Customize colors, sounds, and icons.
5. **Set Your Tiers**:
   - Go to the **Value Tiers** tab. Map your styles to specific price points (e.g., "Use 'High Value' style for anything worth more than 10 Chaos").
6. **Start**: Click **Save & Start AutoFilter**. The program will now monitor the economy and keep your output filter file updated!

## 💡 How It Works

PoEAutoFilter reads your **Base Filter**, appends its own dynamically generated rules at the very top, and saves the combined version to your **Output Filter**.

> [!IMPORTANT]
> When the program updates your file, you still need to **Reload** the filter inside Path of Exile (Escape -> Options -> Game -> Click the 'Reload' button next to your filter) to see the changes in-game.

## ⚙️ Advanced Configuration (Optional)

For users who want even more control, PoEAutoFilter supports:

- **Custom Rules Override**: Directly paste raw filter code into the General tab to have it always appear at the top of your filter.
