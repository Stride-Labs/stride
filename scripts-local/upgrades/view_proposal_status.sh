STRIDED=scripts-local/upgrades/binaries/strided1

$STRIDED query gov proposal 1 | grep "voting_end_time"

while true; do
  $STRIDED query gov proposal 1 | grep "status"
  sleep 5
done
