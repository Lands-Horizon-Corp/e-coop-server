import React, { useEffect, useState } from 'react';
import axios, { type AxiosProgressEvent } from 'axios';
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import { Download } from 'lucide-react';
import { Button } from '../ui/button';
import { TableBody,Table, TableCell, TableHead, TableHeader, TableRow } from '../ui/table';
import { Badge } from '../ui/badge';
import { Progress } from '../ui/progress';
import { useBroadcast } from '@/hook/useBroadcast';
import prettyBytes from 'pretty-bytes';

export type StorageStatus =
    | "pending"
    | "cancelled"
    | "corrupt"
    | "completed"
    | "progress";

export interface Media {
    id: string;
    createdAt: string;
    updatedAt: string;
    fileName: string;
    file_size: number;
    fileType: string;
    storageKey: string;
    url: string;
    key: string;
    downloadURL: string;
    bucketName: string;
    status: StorageStatus;
    progress: number;
}

const statusVariants: Record<Media['status'], string> = {
    pending: 'default',
    progress: 'secondary',
    completed: 'outline',
    cancelled: 'destructive',
    corrupt: 'warning',
};



const SampleMedia: React.FC = () => {
    const [selectedFile, setSelectedFile] = useState<File | null>(null);
    const [uploadProgress, setUploadProgress] = useState<number>(0);
    const [mediaList, setMediaList] = useState<Media[]>([])


    const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>): void => {
        if (e.target.files) {
            setSelectedFile(e.target.files[0]);
        }
    };

    const handleUpload = async (): Promise<void> => {
        if (!selectedFile) return;
        const formData = new FormData();
        formData.append('file', selectedFile);
        try {
            const response = await axios.post(`${import.meta.env.VITE_SERVER_URL}/media`, formData, {
                headers: {
                    'Content-Type': 'multipart/form-data',
                },
                onUploadProgress: (progressEvent: AxiosProgressEvent): void => {
                    if (progressEvent.total) {
                        const progress = Math.round((progressEvent.loaded / progressEvent.total) * 100);
                        setUploadProgress(progress);
                    }
                },
            });
            console.log('Upload successful:', response.data);
        } catch (error) {
            console.error('Error uploading file:', error);
        }
    };
    
    const fetchList = async () => {
        try {
            const res = await axios.get<Media[]>(`${import.meta.env.VITE_SERVER_URL}/media`, { withCredentials: true })
            setMediaList(res.data)
        } catch (error) {
            console.error("List Error:", error)
        }
    }
    useEffect(() => {
        fetchList()
        return () => {
            
        }
    }, [])
      useBroadcast<Media>("media.create", fetchList, console.error)
      useBroadcast<Media>("media.update", fetchList, console.error)
      useBroadcast<Media>("media.delete", fetchList, console.error)
    return (
        <div>
            <div className="grid w-full max-w-sm items-center gap-1.5">
                <Label htmlFor="picture">Picture</Label>
                <Input id="picture" type="file" onChange={handleFileChange} />
                <button onClick={handleUpload} disabled={!selectedFile}>
                    Upload
                </button>
                {uploadProgress > 0 && <p>Progress: {uploadProgress}%</p>}
            </div>
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead>ID</TableHead>
                        <TableHead>File Name</TableHead>
                        <TableHead>Size</TableHead>
                        <TableHead>Type</TableHead>
                        <TableHead>Status</TableHead>
                        <TableHead>Progress</TableHead>
                        <TableHead>Actions</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {mediaList.map((media) => (
                        <TableRow key={media.id}>
                            <TableCell>{media.id}</TableCell>
                            <TableCell>{media.fileName}</TableCell>
                            <TableCell>{prettyBytes(media.file_size)}</TableCell>
                            <TableCell>{media.fileType}</TableCell>
                            <TableCell>
                                <Badge variant={statusVariants[media.status] as any}>
                                    {media.status}
                                </Badge>
                            </TableCell>
                            <TableCell className="w-40">
                                <Progress value={media.progress} />
                                <Label>{media.progress}</Label>
                            </TableCell>
                            <TableCell>
                                {media.downloadURL && (
                                    <a href={media.downloadURL} target="_blank" rel="noopener noreferrer">
                                        <Button variant="ghost" size="sm">
                                            <Download className="mr-1 h-4 w-4" />
                                            Download
                                        </Button>
                                    </a>
                                )}
                            </TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    );
};

export default SampleMedia;
