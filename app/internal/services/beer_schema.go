// Package services provides business logic and service layer functions for Brewsource MCP, including beer and brewery operations.
package services

// SeedBeer represents a sample or seed beer entry with core beer attributes for seeding the database or providing example data.
type SeedBeer struct {
	Name        string
	BreweryName string
	Style       string
	ABV         float64
	IBU         int
	SRM         float64
	Description string
}

// GetSeedBeers returns a slice of sample South African beers.
// Note: The numeric values below are actual beer specifications (ABV, IBU, SRM).
//
//nolint:mnd,funlen // These are real brewing specifications, not magic numbers; dataset function, length is intentional
func GetSeedBeers() []SeedBeer {
	return []SeedBeer{
		// --- SAB (Macro) ---
		{
			Name:        "Castle Lager",
			BreweryName: "SAB - Newlands Brewery",
			Style:       "Pale Lager",
			ABV:         5.0,
			IBU:         18,
			SRM:         3.5,
			Description: "The iconic South African lager, known for its crisp, clean taste and balanced flavour. A national favourite.",
		},
		{
			Name:        "Carling Black Label",
			BreweryName: "SAB - Alrode Brewery",
			Style:       "Pale Lager",
			ABV:         5.5,
			IBU:         20,
			SRM:         4.0,
			Description: "A full-bodied and rewarding lager, famously known as 'Zam-Buk' and celebrated for its champion taste.",
		},
		{
			Name:        "Hansa Pilsner",
			BreweryName: "SAB - Prospecton Brewery",
			Style:       "Pilsner",
			ABV:         4.5,
			IBU:         22,
			SRM:         3.0,
			Description: "A classic pilsner with a distinctive hoppy aroma and a refreshingly crisp finish, brewed with the kiss of the Saaz hop.",
		},
		{
			Name:        "Castle Lite",
			BreweryName: "SAB - Alrode Brewery",
			Style:       "Light Lager",
			ABV:         4.0,
			IBU:         12,
			SRM:         2.5,
			Description: "A premium light lager, extra cold-lagered for a refreshingly crisp and clean taste.",
		},
		{
			Name:        "Castle Milk Stout",
			BreweryName: "SAB - Newlands Brewery",
			Style:       "Milk Stout",
			ABV:         6.0,
			IBU:         28,
			SRM:         40.0,
			Description: "A rich, creamy stout with notes of coffee, chocolate, and dark caramel, offering a smooth and satisfying finish.",
		},
		{
			Name:        "Lion Lager",
			BreweryName: "SAB - Newlands Brewery",
			Style:       "Pale Lager",
			ABV:         4.0,
			IBU:         16,
			SRM:         3.0,
			Description: "A legendary South African lager, rich in heritage, known for its crisp, dry finish.",
		},

		// --- Afro Caribbean Brewing Co. (ACBC) ---
		{
			Name:        "Jungle Paradise",
			BreweryName: "Afro Caribbean Brewing Co. (ACBC)",
			Style:       "Hazy IPA",
			ABV:         6.0,
			IBU:         40,  // Est.
			SRM:         6.0, // Est.
			Description: "A juicy, tropical hazy IPA bursting with flavours of mango, pineapple, and citrus.",
		},
		{
			Name:        "Space Invader",
			BreweryName: "Afro Caribbean Brewing Co. (ACBC)",
			Style:       "American Pale Ale",
			ABV:         5.0,
			IBU:         35,  // Est.
			SRM:         7.0, // Est.
			Description: "A classic APA with a balanced hop and malt profile, featuring citrus and floral notes.",
		},

		// --- Aegir Project Brewery ---
		{
			Name:        "California Steamin'",
			BreweryName: "Aegir Project Brewery",
			Style:       "California Common",
			ABV:         5.0,
			IBU:         35,
			SRM:         12.0,
			Description: "A hybrid lager fermented at ale temperatures, resulting in a malty beer with a rustic hop character.",
		},
		{
			Name:        "Giant's IPA",
			BreweryName: "Aegir Project Brewery",
			Style:       "American IPA",
			ABV:         6.5,
			IBU:         60,  // Est.
			SRM:         8.0, // Est.
			Description: "A big, bold American IPA with a solid hop bitterness and aromas of citrus and pine.",
		},

		// --- Anvil Ale House ---
		{
			Name:        "Anvil Pale Ale",
			BreweryName: "Anvil Ale House",
			Style:       "American Pale Ale",
			ABV:         5.2,
			IBU:         38,  // Est.
			SRM:         9.0, // Est.
			Description: "A balanced and flavourful American Pale Ale, the flagship brew of this Dullstroom establishment.",
		},
		{
			Name:        "White Anvil",
			BreweryName: "Anvil Ale House",
			Style:       "Witbier",
			ABV:         4.8,
			IBU:         15,  // Est.
			SRM:         4.0, // Est.
			Description: "A refreshing Belgian-style wheat beer with notes of coriander and orange peel.",
		},

		// --- Black Horse Brewery & Distillery ---
		{
			Name:        "Black Horse Premium Lager",
			BreweryName: "Black Horse Brewery & Distillery",
			Style:       "Lager",
			ABV:         5.0,
			IBU:         20,  // Est.
			SRM:         4.0, // Est.
			Description: "A clean, crisp, and refreshing lager brewed in the heart of the Magaliesburg.",
		},
		{
			Name:        "Black Horse Weiss",
			BreweryName: "Black Horse Brewery & Distillery",
			Style:       "Weissbier",
			ABV:         4.8,
			IBU:         14,  // Est.
			SRM:         5.0, // Est.
			Description: "A traditional German-style wheat beer with classic banana and clove yeast characteristics.",
		},

		// --- Capital Craft Beer Academy ---
		{
			Name:        "Capital IPA",
			BreweryName: "Capital Craft Beer Academy",
			Style:       "American IPA",
			ABV:         6.0,
			IBU:         55,  // Est.
			SRM:         7.0, // Est.
			Description: "A house IPA known for its solid hop character, often featuring a blend of classic American hops.",
		},

		// --- Cape Brewing Company (CBC) ---
		{
			Name:        "Amber Weiss",
			BreweryName: "Cape Brewing Company (CBC)",
			Style:       "Weissbier",
			ABV:         5.4,
			IBU:         14,
			SRM:         12.0,
			Description: "A classic German-style wheat beer with prominent banana and clove aromas and a smooth, full-bodied mouthfeel.",
		},
		{
			Name:        "Pilsner",
			BreweryName: "Cape Brewing Company (CBC)",
			Style:       "Pilsner",
			ABV:         4.8,
			IBU:         32,
			SRM:         3.0,
			Description: "A crisp and aromatic pilsner, brewed according to the German Purity Law with the finest ingredients.",
		},
		{
			Name:        "Lager",
			BreweryName: "Cape Brewing Company (CBC)",
			Style:       "Lager",
			ABV:         4.8,
			IBU:         18,
			SRM:         4.0,
			Description: "An easy-drinking, clean, and refreshing lager made with 100% malt.",
		},
		{
			Name:        "Harvest Lager",
			BreweryName: "Cape Brewing Company (CBC)",
			Style:       "Kellerbier",
			ABV:         5.0,
			IBU:         22,  // Est.
			SRM:         5.0, // Est.
			Description: "An unfiltered and naturally cloudy lager, offering a fuller malt body and fresh hop aroma.",
		},

		// --- Clarens Brewery ---
		{
			Name:        "Clarens Blonde",
			BreweryName: "Clarens Brewery",
			Style:       "Blonde Ale",
			ABV:         4.5,
			IBU:         15,
			SRM:         4.0,
			Description: "A light, crisp, and refreshing blonde ale. The perfect beer to enjoy in the scenic town of Clarens.",
		},
		{
			Name:        "Clarens English Ale",
			BreweryName: "Clarens Brewery",
			Style:       "English Pale Ale",
			ABV:         5.5,
			IBU:         30,
			SRM:         12.0,
			Description: "A malty, copper-coloured ale with a fruity character, brewed in the traditional English style.",
		},
		{
			Name:        "Clarens IPA",
			BreweryName: "Clarens Brewery",
			Style:       "American IPA",
			ABV:         6.0,
			IBU:         50,
			SRM:         8.0,
			Description: "A classic American IPA with a firm bitterness and aromas of citrus and pine.",
		},

		// --- Darling Brew ---
		{
			Name:        "Bone Crusher",
			BreweryName: "Darling Brew",
			Style:       "Witbier",
			ABV:         5.2,
			IBU:         15,
			SRM:         4.0,
			Description: "A refreshing Belgian-style witbier brewed with coriander and orange peel, named after the Spotted Hyena.",
		},
		{
			Name:        "Slow Beer",
			BreweryName: "Darling Brew",
			Style:       "Lager",
			ABV:         4.0,
			IBU:         22,
			SRM:         3.0,
			Description: "An easy-drinking lager dedicated to the geometric tortoise, offering a crisp taste and a slow, satisfying finish.",
		},
		{
			Name:        "Warlord Imperial IPA",
			BreweryName: "Darling Brew",
			Style:       "Imperial IPA",
			ABV:         9.0,
			IBU:         85,
			SRM:         10.0,
			Description: "A fearsome Imperial IPA with a massive hop profile, inspired by the Martial Eagle. For serious hop lovers.",
		},
		{
			Name:        "Gypsy Mask",
			BreweryName: "Darling Brew",
			Style:       "Red Ale",
			ABV:         4.8,
			IBU:         26,
			SRM:         15.0,
			Description: "A rusty red ale with a smooth, malty character and notes of caramel, inspired by the Roan Antelope.",
		},
		{
			Name:        "Black Mist",
			BreweryName: "Darling Brew",
			Style:       "Black Ale",
			ABV:         5.5,
			IBU:         40,
			SRM:         35.0,
			Description: "A hoppy black ale that combines the roastiness of a stout with the hop profile of an IPA, named for the Verreaux's Eagle.",
		},

		// --- Devil's Peak Brewing Company ---
		{
			Name:        "King's Blockhouse IPA",
			BreweryName: "Devil's Peak Brewing Company",
			Style:       "American IPA",
			ABV:         6.0,
			IBU:         62,
			SRM:         8.0,
			Description: "A bold, hop-forward IPA with citrus and pine notes, regarded as one of South Africa's flagship craft IPAs.",
		},
		{
			Name:        "Devil's Peak Lager",
			BreweryName: "Devil's Peak Brewing Company",
			Style:       "Lager",
			ABV:         4.2,
			IBU:         20,
			SRM:         4.0,
			Description: "A crisp, refreshing lager that is uncomplicated and flavourful. Perfect for any occasion.",
		},
		{
			Name:        "First Light Golden Ale",
			BreweryName: "Devil's Peak Brewing Company",
			Style:       "Golden Ale",
			ABV:         4.0,
			IBU:         18,
			SRM:         5.0,
			Description: "A light and easy-drinking golden ale with a gentle maltiness and subtle fruity hop character.",
		},
		{
			Name:        "The Vannie Hout",
			BreweryName: "Devil's Peak Brewing Company",
			Style:       "Wood-Aged Pale Ale",
			ABV:         6.0,
			IBU:         35,
			SRM:         7.0,
			Description: "A unique pale ale aged on oak, imparting subtle notes of vanilla and spice to complement the hops.",
		},

		// --- Drifter Brewing Company ---
		{
			Name:        "The Stranded Coconut",
			BreweryName: "Drifter Brewing Company",
			Style:       "Coconut Ale",
			ABV:         4.5,
			IBU:         20,
			SRM:         6.0,
			Description: "A unique ale brewed with real coconut, offering a tropical aroma and smooth, refreshing finish.",
		},
		{
			Name:        "Scallywag IPA",
			BreweryName: "Drifter Brewing Company",
			Style:       "IPA",
			ABV:         6.5,
			IBU:         55,
			SRM:         8.0,
			Description: "A classic India Pale Ale with a hoppy punch, showcasing notes of citrus, pine, and a touch of malt sweetness.",
		},
		{
			Name:        "Cape Town Blonde",
			BreweryName: "Drifter Brewing Company",
			Style:       "Blonde Ale",
			ABV:         4.5,
			IBU:         18,  // Est.
			SRM:         4.0, // Est.
			Description: "A light and easy-drinking blonde ale, perfect for the Cape Town lifestyle.",
		},

		// --- Franschhoek Beer Co ---
		{
			Name:        "The Stout",
			BreweryName: "Franschhoek Beer Co",
			Style:       "Oatmeal Stout",
			ABV:         5.2,
			IBU:         30,   // Est.
			SRM:         38.0, // Est.
			Description: "A rich and creamy oatmeal stout with notes of dark chocolate, coffee, and a smooth finish.",
		},
		{
			Name:        "La Saison",
			BreweryName: "Franschhoek Beer Co",
			Style:       "Saison",
			ABV:         6.0,
			IBU:         25,  // Est.
			SRM:         5.0, // Est.
			Description: "A classic Belgian-style saison, offering fruity and spicy yeast notes with a dry, refreshing finish.",
		},
		{
			Name:        "Weiss",
			BreweryName: "Franschhoek Beer Co",
			Style:       "Weissbier",
			ABV:         5.0,
			IBU:         12,
			SRM:         4.0,
			Description: "A traditional German wheat beer with distinct banana and clove esters.",
		},

		// --- Gilroy's Brewery ---
		{
			Name:        "Gilroy's Favourite",
			BreweryName: "Gilroy's Brewery",
			Style:       "Ruby Ale",
			ABV:         5.0,
			IBU:         25,
			SRM:         17.0,
			Description: "A smooth, traditional ruby ale with a fine balance of malt and hops. A true session beer.",
		},
		{
			Name:        "Gilroy's Serious",
			BreweryName: "Gilroy's Brewery",
			Style:       "Old Ale / Dark Ale",
			ABV:         6.5,
			IBU:         40,
			SRM:         25.0,
			Description: "A dark, strong, and seriously flavourful ale for those who appreciate a beer with character and depth.",
		},
		{
			Name:        "Gilroy's Traditional",
			BreweryName: "Gilroy's Brewery",
			Style:       "Premium Lager",
			ABV:         5.0,
			IBU:         22,  // Est.
			SRM:         4.0, // Est.
			Description: "A crisp and clean premium lager made in the traditional style.",
		},

		// --- Hey Joe Brewing Company ---
		{
			Name:        "Belgian Wit",
			BreweryName: "Hey Joe Brewing Company",
			Style:       "Witbier",
			ABV:         5.0,
			IBU:         15,  // Est.
			SRM:         4.0, // Est.
			Description: "A refreshing Belgian Wit brewed with orange peel and coriander, perfect for a sunny day.",
		},
		{
			Name:        "Belgian IPA",
			BreweryName: "Hey Joe Brewing Company",
			Style:       "Belgian IPA",
			ABV:         6.5,
			IBU:         50,  // Est.
			SRM:         7.0, // Est.
			Description: "A hybrid style combining the spicy, fruity notes of Belgian yeast with the bitterness of an American IPA.",
		},

		// --- Jack Black's Brewing Company ---
		{
			Name:        "Brewers Lager",
			BreweryName: "Jack Black's Brewing Company",
			Style:       "Lager",
			ABV:         5.0,
			IBU:         22,
			SRM:         4.0,
			Description: "A flagship, all-malt lager that is crisp, clean, and well-balanced with a satisfying malt backbone.",
		},
		{
			Name:        "Cape Pale Ale (CPA)",
			BreweryName: "Jack Black's Brewing Company",
			Style:       "American Pale Ale",
			ABV:         5.5,
			IBU:         35,
			SRM:         9.0,
			Description: "A vibrant pale ale with layered citrus and floral hop aromas, delivering a refreshing and flavourful experience.",
		},
		{
			Name:        "Skeleton Coast IPA",
			BreweryName: "Jack Black's Brewing Company",
			Style:       "American IPA",
			ABV:         6.5,
			IBU:         60,
			SRM:         7.0,
			Description: "An assertive, hop-driven IPA bursting with tropical fruit notes and a solid malt foundation to balance the bitterness.",
		},
		{
			Name:        "Atlantic Weiss",
			BreweryName: "Jack Black's Brewing Company",
			Style:       "Weissbier",
			ABV:         5.0,
			IBU:         14,  // Est.
			SRM:         5.0, // Est.
			Description: "A refreshing, unfiltered wheat beer with low bitterness and classic notes of banana and clove.",
		},

		// --- Mad Giant Brewery ---
		{
			Name:        "Urban Legend IPA",
			BreweryName: "Mad Giant Brewery",
			Style:       "American IPA",
			ABV:         6.0,
			IBU:         55,
			SRM:         8.0,
			Description: "A hop-forward IPA that balances bitterness with a strong malt backbone, featuring notes of citrus and tropical fruit.",
		},
		{
			Name:        "Killer Hop",
			BreweryName: "Mad Giant Brewery",
			Style:       "Pale Ale",
			ABV:         5.0,
			IBU:         35,
			SRM:         7.0,
			Description: "A refreshingly crisp and aromatic pale ale with a killer combination of hops, delivering a hoppy but sessionable beer.",
		},
		{
			Name:        "Mad Giant Pilsner",
			BreweryName: "Mad Giant Brewery",
			Style:       "Pilsner",
			ABV:         4.2,
			IBU:         25,  // Est.
			SRM:         3.0, // Est.
			Description: "A crisp, clean, and refreshing pilsner with a noble hop character.",
		},

		// --- Richmond Hill Brewing Co ---
		{
			Name:        "Car Park John",
			BreweryName: "Richmond Hill Brewing Co",
			Style:       "American Pale Ale",
			ABV:         5.0,
			IBU:         38,  // Est.
			SRM:         8.0, // Est.
			Description: "A flagship APA with a delightful balance of citrusy hops and biscuit malt.",
		},
		{
			Name:        "Two Rand Man",
			BreweryName: "Richmond Hill Brewing Co",
			Style:       "Lager",
			ABV:         4.5,
			IBU:         18,  // Est.
			SRM:         4.0, // Est.
			Description: "An easy-drinking, crisp, and clean lager for any occasion.",
		},

		// --- Saggy Stone Brewing Co. ---
		{
			Name:        "California Steam",
			BreweryName: "Saggy Stone Brewing Co.",
			Style:       "California Common",
			ABV:         5.0,
			IBU:         35,   // Est.
			SRM:         11.0, // Est.
			Description: "A smooth, malty beer with a distinct hop character, brewed in the unique California Common style.",
		},
		{
			Name:        "Rocky River Lager",
			BreweryName: "Saggy Stone Brewing Co.",
			Style:       "Lager",
			ABV:         4.5,
			IBU:         20,  // Est.
			SRM:         4.0, // Est.
			Description: "A crisp and refreshing lager, perfect for enjoying in the scenic Nuy Valley.",
		},

		// --- Soul Barrel Brewing Co. ---
		{
			Name:        "Live Culture",
			BreweryName: "Soul Barrel Brewing Co.",
			Style:       "American Pale Ale",
			ABV:         5.5,
			IBU:         35,
			SRM:         6.0,
			Description: "A farmhouse-style pale ale fermented with wild yeast, offering complex fruity and earthy notes.",
		},
		{
			Name:        "Cape Cone",
			BreweryName: "Soul Barrel Brewing Co.",
			Style:       "South African IPA",
			ABV:         6.5,
			IBU:         60,
			SRM:         7.0,
			Description: "An IPA brewed exclusively with locally grown South African hops, showcasing unique flavours of tropical fruit and citrus.",
		},
		{
			Name:        "Oud Bruin",
			BreweryName: "Soul Barrel Brewing Co.",
			Style:       "Flanders Oud Bruin",
			ABV:         7.0,
			IBU:         20,   // Est.
			SRM:         18.0, // Est.
			Description: "A complex, barrel-aged sour ale with notes of dark fruit, malt, and a pleasant tartness.",
		},

		// --- Stellenbosch Brewing Company ---
		{
			Name:        "Hoenderhok Bock",
			BreweryName: "Stellenbosch Brewing Company",
			Style:       "Bock",
			ABV:         6.5,
			IBU:         24,
			SRM:         20.0,
			Description: "A malty, rich bock with caramel and toffee notes, inspired by German brewing traditions.",
		},
		{
			Name:        "Eikeboom Helles",
			BreweryName: "Stellenbosch Brewing Company",
			Style:       "Helles Lager",
			ABV:         4.5,
			IBU:         18,
			SRM:         3.5,
			Description: "A clean, crisp, and easy-drinking German-style Helles, perfect for a sunny Stellenbosch day.",
		},
		{
			Name:        "Bosch Weiss",
			BreweryName: "Stellenbosch Brewing Company",
			Style:       "Weissbier",
			ABV:         5.0,
			IBU:         14,  // Est.
			SRM:         5.0, // Est.
			Description: "A classic hefeweizen with notes of banana and clove from the traditional yeast strain.",
		},

		// --- That Brewing Company ---
		{
			Name:        "That Blonde",
			BreweryName: "That Brewing Company",
			Style:       "Blonde Ale",
			ABV:         4.5,
			IBU:         20,  // Est.
			SRM:         4.0, // Est.
			Description: "A light, sessionable blonde ale that is clean and crisp.",
		},
		{
			Name:        "That Good Ad Weiss",
			BreweryName: "That Brewing Company",
			Style:       "Weissbier",
			ABV:         5.0,
			IBU:         15,  // Est.
			SRM:         5.0, // Est.
			Description: "A refreshing hefeweizen with the typical banana and clove characteristics.",
		},

		// --- Woodstock Brewery ---
		{
			Name:        "Californicator IPA",
			BreweryName: "Woodstock Brewery",
			Style:       "West Coast IPA",
			ABV:         6.5,
			IBU:         65,
			SRM:         7.0,
			Description: "A classic West Coast IPA with aggressive hopping, delivering big citrus and pine flavours with a dry finish.",
		},
		{
			Name:        "Pot Belge",
			BreweryName: "Woodstock Brewery",
			Style:       "Belgian Pale Ale",
			ABV:         7.0,
			IBU:         25,
			SRM:         6.0,
			Description: "A Belgian-style ale with characteristic fruity and spicy yeast notes, balanced by a pleasant malt profile.",
		},
		{
			Name:        "Happy Pills Pilsner",
			BreweryName: "Woodstock Brewery",
			Style:       "Pilsner",
			ABV:         5.0,
			IBU:         30,
			SRM:         3.5,
			Description: "A clean, crisp, and refreshing pilsner that's perfect for any occasion.",
		},

		// --- 1000 Hills Brewing Company ---
		{
			Name:        "The Dean",
			BreweryName: "1000 Hills Brewing Company",
			Style:       "English Pale Ale",
			ABV:         4.5,
			IBU:         30,   // Est.
			SRM:         10.0, // Est.
			Description: "A traditional English-style Pale Ale with a solid malt backbone and earthy hop notes.",
		},
		{
			Name:        "The Graduate",
			BreweryName: "1000 Hills Brewing Company",
			Style:       "American Pale Ale",
			ABV:         5.5,
			IBU:         38,  // Est.
			SRM:         8.0, // Est.
			Description: "A hop-forward American Pale Ale with bright citrus and floral aromas.",
		},

		// --- The Kennel Brewery (New Addition) ---
		{
			Name:        "Aleprávda",
			BreweryName: "The Kennel Brewery",
			Style:       "American Pale Ale",
			ABV:         5.5,
			IBU:         40,
			SRM:         9.0,
			Description: "A single-hop (Cascade) APA that is exceptionally well-balanced and highly sessionable.",
		},

		// --- Valley of Darkness (New Addition) ---
		{
			Name:        "Black Dog",
			BreweryName: "Valley of Darkness",
			Style:       "Belgian Dark Strong Ale",
			ABV:         9.5,
			IBU:         30,   // Est.
			SRM:         25.0, // Est.
			Description: "A complex and strong Belgian ale with notes of dark fruit, caramel, and spice. Not for the faint-hearted.",
		},
	}
}
