clear
reset
set term png size 1800, 1200

set yrange [0:*]

# Format the y-axis ticks to display as dollars
set format y "$%'g M"
set decimal locale

set output "output.png"
plot "output.dat" using 1:($2 * 0.00000001) with lines title "Brokerage", \
     "output.dat" using 1:($3 * 0.00000001) with lines title "IRA", \
     "output.dat" using 1:($4 * 0.00000001) with lines title "Expenses"