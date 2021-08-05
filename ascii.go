package main

import (
	"fmt"
	"strings"
	"time"
)

func printFunnelArt(withAnimation ...bool) {
	art := `

                ,-""-.     ,-""-.     ,-""-.
               / ,--. \   / ,--. \   / ,--. \
              | ( () ) | | ( () ) | | ( () ) |
               \ '--' /   \ '--' /   \ '--' /
                '-..-'     '-..-'     '-..-'

        @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
         @                                        @(
         .@                 BAIT                 @@
          %@                                    @@
           @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
            @(                                 @
            %@            FRONTEND            @(
             @                                @
              @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@.
               @                            @(
               .@          MIDDLE          @@
                %@                        &@
                 @@@@@@@@@@@@@@@@@@@@@@@@@@
                  @(                     @
                   @       BACKEND      @
                    @                  @.
                     @@@@@@@@@@@@@@@@@@#
                        $     $    $
                       $$$   $$$  $$$
                      $$$$$ $$$$$ $$$$

   `

	if len(withAnimation) == 0 {
		fmt.Println(art)
		return
	}
	if !withAnimation[0] {
		fmt.Println(art)
		return
	}

	arr := strings.Split(art, "\n")
	for _, line := range arr {
		time.Sleep(time.Millisecond * 25)
		fmt.Println(line)
	}
}
