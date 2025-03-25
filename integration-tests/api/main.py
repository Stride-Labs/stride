from fastapi import FastAPI, File, HTTPException, UploadFile
from fastapi.responses import FileResponse
import os

app = FastAPI()

STORAGE_DIRECTORY = "storage"
os.makedirs(STORAGE_DIRECTORY, exist_ok=True)


@app.get("/status")
async def status() -> str:
    """
    Health check
    """
    return "ok"


@app.post("/upload/{file_name}")
@app.post("/upload/{path:path}/{file_name}")
async def upload_file(file_name: str, path: str = "", file: UploadFile = File(...)) -> dict:
    """
    Allows uploading a file - stores it in the local file system
    """
    parent = f"{STORAGE_DIRECTORY}/{path}" if path else STORAGE_DIRECTORY
    os.makedirs(parent, exist_ok=True)

    with open(f"{parent}/{file_name}", "wb") as f:
        f.write(file.file.read())

    return {"info": f"file {file_name} saved"}


@app.get("/download/{file_name}")
@app.get("/download/{path:path}/{file_name}")
async def download_file(file_name: str, path: str = ""):
    """
    Allows downloading a file from the local file system
    """
    file_location = f"{STORAGE_DIRECTORY}/{path}/{file_name}" if path else f"{STORAGE_DIRECTORY}/{file_name}"

    if not os.path.exists(file_location):
        return HTTPException(status_code=404, detail=f"file {file_location} not found")

    return FileResponse(file_location, media_type="application/octet-stream", filename=file_name)
