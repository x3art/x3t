package xt

func GetLasers(n string, text Text) []Laser {
	ret := []Laser{}
	tparse(n, text, &ret)
	return ret
}

type Laser struct {
	BodyFile               string
	PictureID              string
	RotX                   float64
	RotY                   float64
	RotZ                   float64
	Index                  string
	Description            string `x3t:"page:17"`
	RoF                    int
	Sound                  int
	Projectile             int
	Energy                 int
	ChargeRate             float64
	HUDIcon                string
	Volume                 string
	ProductionRelValNPC    string
	PriceMod1              string
	PriceMod2              string
	WareClass              string
	ProductionRelValPlayer string
	MinNotoriety           string
	VideoID                string
	Skin                   string
	OjectID                string
}

/*
Field	Description
0. Body file	Body displayed when the weapon is floating in space
1. Picture ID	Not used
2. Rotation X	Vertical turret speed
3. Rotation Y	Horizontal turret speed
4. Rotation Z	Not used
5. Index
6. Description	String ID from Page 17 of text resource files
7. Rate of fire	The value is 1 shot per x milliseconds (1000 = 60 rounds per minute)
8. Sound	Index to Sounds.txt
9. Projectile	Projectile fired by this weapon - index to TBullets.txt
10. Energy	How much energy does the weapon have
11. Charge rate	In percents per second (0.5 = 50% per second will be restored)
12. HUD Icon	Icon name from IconData.txt
13. Volume
14. Production RelVal (NPC)	Price for NPCs (it's not really a price)
15. Price modifier PRI	Primary Price Modifier
16. Price modifier SEC	Secondary Price Modifier
17. Ware class	Class (size) of the object - affects which ships can carry it
18. Production RelVal (player)	Price for the player (it's not really a price)
19. Min. Notoriety	Minimal notoriety the player must have to be able to use this object
20. Video ID	Stream ID from Videos.txt containing the animation displayed in the Info screen
21. Skin	Index to Skins.txt
22. Object ID	Identifier of the object
*/
