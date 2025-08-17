# PoEAutoFilter

This program is designed to automatically print Path Of Exile filter rules based on a snapshot of the economy. While running, the program automatically updates your filter hourly to the *current* economy. This program cannot manually push your client to update to the current filter file . It will load the current filter file when you open the client or when manually clicking the reload file button in the game settings tab.

Currently, the program fetches data from poe.ninja while running, gets the current Chaos, Exalted, and Divine values, and calculates how big of a stack of each supported item would need to drop in order to meet certain thresholds. The program is bracketed into value tiers of: Sub 1 Chaos(this tier is modulated by a multiplier value set in the config. You can choose to scale the bottom floor of the filter from .1 to 1 chaos to set your minimum value threshold for what should show), 1 Chaos, 5 Chaos, 1 Exalted, and 0.5 Divine. For instance, with the economy at time of writing, Glassblower Baubles are valued roughly 4 to 1 Chaos. So in order to trigger the 1 Chaos value tier of the filter, a stack of Baubles would need at least 4. On my filter, i have the Sub1cMult config setting at 1, so i will never show a stack of Baubles less than 4, but if your mult is set to 0.5 for instance, you would see stacks of 2 Baubles.

The filter **ONLY** supports select item types, currently these are Currencies, Fragments, Scarabs, Essences, and Fossils. You will need a base filter to manipulate, with rules for all other items. I suggest using [filterblade](https://www.filterblade.xyz) to generate one.

## SETUP

In order to setup the program you will need to download the .exe and setup your config file. In the future, i would like to implement a GUI to remove the need to manually edit a text file. For now, setup the config file as such:

1. **FilePath=**

Here you will need to locate your filter file folder. You can do this ingame by navigating to the game options tab and clicking the folder button in the item filter line. Once the folder is opened, you can click in the url bar at the top to get the file path. Make sure to include the file name at the end. For instance, my config.txt FilePath= line would look like this:

FilePath= C:\Users\Username\Documents\My Games\Path of Exile\AutoFilter.filter

2. **League=** 

Here you will need to set the current League you are playing in whose economy the program will base its valuations on. For instance, this program was developed during Mercenaries of Trarthus league, so the filter line would look like this:

League= Mercenaries

However, some leagues have odd naming conventions, so in case it isn't obvious what the name of the trade league will be, you can find this by navigating to the [trade website](https://www.pathofexile.com/trade) and find the league you want to use in the dropdown menu next to the search bar. Once you select it, you should see the name of the league in the URL. Mine says "https://www.pathofexile.com/trade/search/Mercenaries", so Mercenaries is the name i need to use for this league.

3. **Sub1cMult=**

For this line, you will set the floor of what value a stack of items would need in order to not be hidden. You can set any value between 0 and 1 here. For now, the floor threshold cannot go above 1 so Chaos orbs will never be hidden. Example:

Sub1cMult= 0.5

4. **Running the program the first time**

This program ***ONLY*** supports certain stackable currency items, so you are expected to provide a basic filter that will have rules for everything else such as rare items and such. You can create your own base file using [filterblade](https://www.filterblade.xyz) for instance. I named this file "AutoFilter.filter" and dragged it to the Path Of Exile filter folder, then set the FilePath= line to reflect the file's location. You can at any time, replace the file with a new base and just restart the program to add the current economy rules to the new base. Once you have a base filter and have set the config file correctly, you can run the .exe file. A command prompt window will open to show you the status. Each hour, the program will update and print the results. It should look something like this:

`Hello, Path of Exile Auto Filter!`

`Fetching item values for currency type: Currency`

`Found 111 items`

`Current Prices:`

`Chaos Orb: 1.000000c`

`Exalted Orb: 22.780000c`

`Divine Orb: 133.790000c`

`Fetching item values for fragments type: Fragment`

`Found 80 fragments`

`Fetching item values for scarabs type: Scarab`

`Found 109 scarabs`

`Fetching item values for fossils type: Fossil`

`Found 25 fossils`

`Fetching item values for essences type: Essence`

`Found 105 essences`

`Filter file updated successfully!`

`Filter blocks written to file: C:\Users\Username\Documents\My Games\Path of Exile\AutoFilter.filter`

`At Time: 2025-08-16 18:32:23.7810086 -0700 PDT m=+0.122500301`

`Waiting 1 hour before next update...`

You can leave this running in order to keep your file as up to date as possible, or run it once each time you get on to play, before starting the game. Closing the terminal window will stop the program, minimizing it will allow it to run in the background. Remember that you **MUST** remember to manually update the file ingame for it to stay updated if you are playing for a long period. Otherwise the game client will only read the file when you login to a character using that filter.