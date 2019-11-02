package bot

import "testing"

// s is the test version of the leaderboard.
var s string = `:cityscape:  |  Guild Score Leaderboards for hehe central

	📋 Rank | Name
	
	[1]     > #ode
				Total Score: 156661    
	[2]     > #Crouton
				Total Score: 76926     
	[3]     > #Pupper
				Total Score: 58203     
	[4]     > #theBeef
				Total Score: 45119     
	[5]     > #dylwalk
				Total Score: 28147     
	[6]     > #Sirobot
				Total Score: 21648     
	[7]     > #Washingdone
				Total Score: 21194     
	[8]     > #KyroZed
				Total Score: 17812     
	[9]     > #thy
				Total Score: 17567     
	[10]    > #Naldy
				Total Score: 13214     
	-------------------------------------
	# Your Guild Placing Stats
	😐 Rank: 2    Total Score: 76926`

func TestParse(t *testing.T) {
	results, err := Parse(s)
	if err != nil {
		t.Error(err)
	}
	t.Log(results)
}
