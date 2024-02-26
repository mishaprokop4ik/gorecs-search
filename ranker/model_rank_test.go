package ranker_test

import (
	"github.com/mishaprokop4ik/gorecs-search/lexer"
	"github.com/mishaprokop4ik/gorecs-search/ranker"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestModel_Rank(t *testing.T) {
	content1 := `William Shakespeare's name is synonymous with many of the famous lines he wrote in his plays and prose. Yet his poems are not nearly as recognizable to many as the characters and famous monologues from his many plays.

In Shakespeare's era (1564-1616), it was not profitable but very fashionable to write poetry. It also provided credibility to his talent as a writer and helped to enhance his social standing. It seems writing poetry was something he greatly enjoyed and did mainly for himself at times when he was not consumed with writing a play. Because of their more private nature, few poems, particularly long-form poems, have been published.

The two longest works William Shakespeare that scholars agree were written by Shakespeare are entitled Venus and Adonis and The Rape of Lucrece. Both dedicated to the Honorable Henry Wriothesley, Earl of Southampton, who seems to have acted as a sponsor and encouraging benefactor of Shakespeare's work for a brief time.

Both of these poems contain dozens of stanzas and comment on the depravity of unwanted sexual advances, showing themes throughout of guilt, lust, and moral confusion. In Venus and Adonis, an innocent Adonis must reject the sexual advances of Venus. Conversely in The Rape of Lucrece, the honorable and virtuous wife Lucrece is raped a character overcome with lust, Tarquin. The dedication to Wriothesley is much warmer in the second poem, suggesting a deepening of their relationship and Shakespeare's appreciation of his support.

A third and shorter narrative poem, William A Lover's Complaint, was printed in the first collection of Shakespeare's sonnets. Most scholars agree now that it was also written by Shakespeare, though that was contested for some time. The poem tells the story of a young woman who is driven to misery by a persuasive suitor's attempts to seduce her. It is not regarded by critics to be his finest work.

Another short poem, The Phoenix and the Turtle, despairs the death of a legendary phoenix and his faithful turtle dove lover. It speaks to the frailty of love and commitment in a world where only death is certain.

There are 152 short sonnets attributed to Shakespeare. Among them, the most famous ones are Sonnet 29, Sonnet 71, and Sonnet 55. As a collection, narrative sequence of his Sonnets speaks to Shakespeare's deep insecurity and jealousy as a lover, his grief at separation, and his delight in sharing beautiful experiences with his romantic counterparts. However, few scholars believe that the sequence of the sonnets accurately depicts the order in which they were written. Because Shakespeare seemed to write primarily for his own private audience, dating these short jewels of literature has been next to impossible.

Within the sonnets Shakespeare seems to have two deliberate series: one describing his all consuming lust for a married woman with a dark complexion (the Dark Lady), and one about his confused love feelings for a handsome young man (the Fair Youth). This dichotomy has been widely studied and debated and it remains unclear as to if the subjects represented real people or two opposing sides to Shakespeare's own personality.

Though some of Shakespeare's poetry was published without his permission in his lifetime, in texts such as The Passionate Pilgrim, the majority of the sonnets were published in 1609 by Thomas Thorpe. Before that time, it appears that Shakespeare would only have shared his poetry with a very close inner-circle of friends and loved ones. Thorpe's collection was the last of Shakespeare's non-dramatic work to be printed before his death.`

	l := lexer.NewLexer(content1)
	parsedInfo1 := []string{}
	for term := l.Next(); term != ""; term = l.Next() {
		parsedInfo1 = append(parsedInfo1, term)
	}

	content2 := `test All’s Well That Ends Well, comedy in five acts by Shakespeare, written in 1601–05 and published in the First Folio of 1623 seemingly from a theatrical playbook that still retained certain authorial features or from a literary transcript either of the playbook or of an authorial manuscript. The principal source of the plot was a tale in Giovanni Boccaccio’s Decameron.

The play concerns the test efforts of Helena, daughter of a renowned physician to the recently deceased count of Rossillion, to win as her husband the young new count, Bertram. When Bertram leaves Rossillion to become a courtier, Helena follows after, hoping to minister to the gravely ill king of France with a miraculous cure that her father had bequeathed to her. In return for her success in doing so, the king invites her to select a husband, her choice being Bertram. The young man, unwilling to marry so far below himself in social station, accedes to the royal imperative but promptly flees to military action in Tuscany with his vapid but engaging friend Parolles. By letter Bertram informs Helena that he may not be considered her husband until she has taken the ring from his finger and conceived a child by him. Disguised as a pilgrim, Helena follows Bertram to Florence only to discover that he has been courting Diana, the daughter of her hostess. Helena spreads a rumour of her own death and arranges a nighttime rendezvous with Bertram in which she substitutes herself for Diana. In exchange for his ring, she gives him one that the king has given her. When Bertram returns to Rossillion, where the king is visiting the countess, the royal guest recognizes the ring and suspects foul play. Helena then appears to explain her machinations and claim her rightful spouse.`

	l = lexer.NewLexer(content2)
	parsedInfo2 := []string{}
	for term := l.Next(); term != ""; term = l.Next() {
		parsedInfo2 = append(parsedInfo2, term)
	}
	m := ranker.NewModel(map[string][]string{
		"local1": parsedInfo1,
		"local2": parsedInfo2,
	})

	res := m.Rank("william", "shakespeare", "to")

	assert.Equal(t, []ranker.Path{"local1"}, res)
}
