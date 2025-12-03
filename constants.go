package main

// --- Constants ---
const (
	ScreenWidth      = 1280
	ScreenHeight     = 720
	PlayerSize       = 128
	PlayerSpeed      = 5.0
	ScrollSpeed      = 4.0  // Speed at which the environment scrolls (increased from 3.0)
	ObstacleMinGap   = 350   // Minimum vertical space for the path (decreased from 400)
	ObstacleMaxGap   = 350   // Maximum vertical space for the path (decreased from 400)
	NumFish          = 14    // Number of fish following the leader
	FishSize         = 48    // Size of each following fish
	FishFollowSpeed  = 4.0   // Speed at which fish follow
	CircleRadius     = 100.0 // Radius of the circle behind the leader
	CircleOffsetX    = -100.0 // X offset of the circle center behind the leader
	PlayerX          = 200.0  // Fixed X position of the leader
	FishWanderRadius = 40.0   // Radius within which fish can wander from their base position
	FishWanderIntervalMin = 60   // Minimum frames between wander target changes (1 second at 60 FPS)
	FishWanderIntervalMax = 180  // Maximum frames between wander target changes (3 seconds at 60 FPS)
	NumBackgroundFish = 8     // Number of background ambient fish
	NumBubbles        = 15    // Number of floating bubbles
)

