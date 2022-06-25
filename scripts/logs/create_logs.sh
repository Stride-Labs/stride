SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# clean up logs
rm -f $SCRIPT_DIR/*.log

docker compose logs >> $SCRIPT_DIR/all.log
docker compose logs stride1 >> $SCRIPT_DIR/stride.log
docker compose logs gaia1 >> $SCRIPT_DIR/gaia.log
docker compose logs icq >> $SCRIPT_DIR/icq.log
docker compose logs hermes >> $SCRIPT_DIR/hermes.log